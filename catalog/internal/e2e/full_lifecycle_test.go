package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	mysqlDriver "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testContainer "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/vterry/food-ordering/catalog/internal/adapters/input/rest"
	"github.com/vterry/food-ordering/catalog/internal/adapters/output/repository"
	"github.com/vterry/food-ordering/catalog/internal/core/app"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/services"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
)

var (
	testDB      *sql.DB
	dbContainer *testContainer.MySQLContainer
	testMux     *http.ServeMux
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	dbContainer, err = testContainer.Run(ctx, "mysql:8.0",
		testContainer.WithDatabase("catalog_e2e"),
		testContainer.WithUsername("root"),
		testContainer.WithPassword("test"))

	if err != nil {
		log.Fatalf("failed to start db container: %v", err)
	}

	connString, err := dbContainer.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	testDB, err = sql.Open("mysql", connString)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	runMigrations(testDB)
	testMux = wireApplication(testDB)

	code := m.Run()

	testDB.Close()
	dbContainer.Terminate(ctx)
	os.Exit(code)
}

func wireApplication(db *sql.DB) *http.ServeMux {
	logger := slog.Default()

	outboxRepo := repository.NewOutboxRepository(db)
	menuRepo := repository.NewMenuRepository(db, outboxRepo)
	restaurantRepo := repository.NewRestaurantRepository(db, outboxRepo)
	unitOfWork := repository.NewUnitOfWork(db)

	assignMenuService := services.NewMenuAssignmentService()

	menuAppService := app.NewMenuAppService(assignMenuService, unitOfWork, menuRepo, restaurantRepo)
	restaurantAppService := app.NewRestaurantAppService(unitOfWork, restaurantRepo)

	menuHandler := rest.NewMenuHandler(menuAppService, logger)
	restaurantHandler := rest.NewRestaurantHandler(restaurantAppService, logger)

	mux := http.NewServeMux()
	menuHandler.RegisterRoutes(mux)
	restaurantHandler.RegisterRoutes(mux)

	return mux
}

func runMigrations(db *sql.DB) {
	driver, err := mysqlDriver.WithInstance(db, &mysqlDriver.Config{})
	if err != nil {
		log.Fatalf("could not create driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../infra/db/migrate/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run migrations: %v", err)
	}
}

func truncateTables(db *sql.DB) {
	tables := []string{"items", "categories", "menus", "outbox_events", "restaurants"}
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	for _, table := range tables {
		if _, err := db.Exec("TRUNCATE TABLE " + table); err != nil {
			log.Fatalf("failed to truncate %s: %v", table, err)
		}
	}
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
}

// ===== Helper functions =====

func doPost(t *testing.T, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testMux.ServeHTTP(rec, req)
	return rec
}

func doPatch(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPatch, path, nil)
	rec := httptest.NewRecorder()
	testMux.ServeHTTP(rec, req)
	return rec
}

func doGet(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	testMux.ServeHTTP(rec, req)
	return rec
}

func parseResponse(t *testing.T, rec *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	err := json.NewDecoder(rec.Body).Decode(target)
	require.NoError(t, err)
}

// ===== E2E Test =====

func TestE2E_FullMenuLifecycle(t *testing.T) {
	truncateTables(testDB)

	// ========================================
	// Step 1: Create Restaurant
	// ========================================
	t.Log("Step 1: Creating restaurant...")

	createRestReq := input.CreateRestaurantRequest{
		Name: "Pizzaria E2E",
		Address: input.AddressRequest{
			Street: "Rua dos Testes", Number: "42", Neighborhood: "Centro",
			City: "São Paulo", State: "SP", ZipCode: "01000000",
		},
	}

	rec := doPost(t, "/restaurant", createRestReq)
	require.Equal(t, http.StatusCreated, rec.Code, "create restaurant should return 201")

	var restResp input.RestaurantResponse
	parseResponse(t, rec, &restResp)

	assert.NotEmpty(t, restResp.ID)
	assert.Equal(t, "Pizzaria E2E", restResp.Name)
	assert.Equal(t, "CLOSED", restResp.Status)

	restaurantID := restResp.ID
	t.Logf("  Restaurant created: %s", restaurantID)

	// ========================================
	// Step 2: Create Menu for the Restaurant
	// ========================================
	t.Log("Step 2: Creating menu...")

	createMenuReq := input.CreateMenuRequest{Name: "Menu Principal"}
	rec = doPost(t, "/restaurant/"+restaurantID+"/menu", createMenuReq)
	require.Equal(t, http.StatusCreated, rec.Code, "create menu should return 201")

	var menuResp input.MenuResponse
	parseResponse(t, rec, &menuResp)

	assert.NotEmpty(t, menuResp.ID)
	assert.Equal(t, "Menu Principal", menuResp.Name)
	assert.Equal(t, restaurantID, menuResp.RestaurantID)
	assert.Equal(t, "DRAFT", menuResp.Status)

	menuID := menuResp.ID
	t.Logf("  Menu created: %s (status: %s)", menuID, menuResp.Status)

	// ========================================
	// Step 3: Add Category to Menu
	// ========================================
	t.Log("Step 3: Adding category...")

	addCatReq := input.AddCategoryRequest{Name: "Pizzas Tradicionais"}
	rec = doPost(t, "/menu/"+menuID+"/categories", addCatReq)
	require.Equal(t, http.StatusCreated, rec.Code, "add category should return 201")

	t.Log("  Category 'Pizzas Tradicionais' added")

	// ========================================
	// Step 4: Get Menu to find the category ID
	// ========================================
	t.Log("Step 4: Fetching category ID from database...")

	var categoryID string
	err := testDB.QueryRow(
		"SELECT uuid FROM categories WHERE menu_id = (SELECT id FROM menus WHERE uuid = ?)",
		menuID,
	).Scan(&categoryID)
	require.NoError(t, err, "category should exist in DB")

	t.Logf("  Category ID: %s", categoryID)

	// ========================================
	// Step 5: Add Item to Category
	// ========================================
	t.Log("Step 5: Adding item to category...")

	addItemReq := input.AddItemRequest{
		Name:        "Margherita",
		Description: "Molho de tomate, mozzarella, manjericão fresco",
		PriceCents:  3500,
	}
	rec = doPost(t, "/menu/"+menuID+"/categories/"+categoryID+"/item", addItemReq)
	require.Equal(t, http.StatusCreated, rec.Code, "add item should return 201")

	t.Log("  Item 'Margherita' added (R$35,00)")

	// ========================================
	// Step 6: Activate Menu
	// ========================================
	t.Log("Step 6: Activating menu...")

	rec = doPatch(t, "/menu/"+menuID+"/activate")
	require.Equal(t, http.StatusOK, rec.Code, "activate menu should return 200")

	t.Log("  Menu activated")

	// ========================================
	// Step 7: Get Active Menu via REST
	// ========================================
	t.Log("Step 7: Fetching active menu...")

	rec = doGet(t, "/restaurant/"+restaurantID+"/menu")
	require.Equal(t, http.StatusOK, rec.Code, "get active menu should return 200")

	var activeMenuResp input.MenuResponse
	parseResponse(t, rec, &activeMenuResp)

	assert.Equal(t, menuID, activeMenuResp.ID)
	assert.Equal(t, "Menu Principal", activeMenuResp.Name)
	assert.Equal(t, "ACTIVE", activeMenuResp.Status)
	assert.Equal(t, restaurantID, activeMenuResp.RestaurantID)

	// Validate categories and items
	require.Len(t, activeMenuResp.Categories, 1, "should have 1 category")
	assert.Equal(t, "Pizzas Tradicionais", activeMenuResp.Categories[0].Name)

	require.Len(t, activeMenuResp.Categories[0].Items, 1, "should have 1 item")
	assert.Equal(t, "Margherita", activeMenuResp.Categories[0].Items[0].Name)
	assert.Equal(t, int64(3500), activeMenuResp.Categories[0].Items[0].PriceCents)
	assert.Equal(t, "AVAILABLE", activeMenuResp.Categories[0].Items[0].Status)

	t.Logf("  Active menu verified: %s with %d categories, %d items",
		activeMenuResp.Name,
		len(activeMenuResp.Categories),
		len(activeMenuResp.Categories[0].Items),
	)

	// ========================================
	// Step 8: Verify Restaurant has assignment
	// ========================================
	t.Log("Step 8: Verifying restaurant has active menu assigned...")

	rec = doGet(t, "/restaurant/"+restaurantID)
	require.Equal(t, http.StatusOK, rec.Code, "get restaurant should return 200")

	var updatedRestResp input.RestaurantResponse
	parseResponse(t, rec, &updatedRestResp)

	assert.Equal(t, menuID, updatedRestResp.ActiveMenuID, "restaurant should have active menu assigned")
	assert.Equal(t, "CLOSED", updatedRestResp.Status) // still closed — Open requires explicit PATCH

	t.Logf("  Restaurant has active_menu_id: %s", updatedRestResp.ActiveMenuID)

	// ========================================
	// Step 9: Verify outbox events were created
	// ========================================
	t.Log("Step 9: Verifying outbox events...")

	var eventCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM outbox_events").Scan(&eventCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, eventCount, 3, "should have at least 3 outbox events (RestaurantCreated, MenuCreated, MenuActivated, RestaurantMenuUpdated)")

	t.Logf("  Total outbox events: %d", eventCount)

	t.Log("✅ E2E Full Menu Lifecycle — PASSED!")
}
