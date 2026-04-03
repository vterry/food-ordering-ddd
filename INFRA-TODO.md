# 🚀 Infraestrutura K8s / Istio - Próximos Passos (TODO)

Este documento registra o estado atual e os próximos passos para a configuração da infraestrutura de desenvolvimento local (Kubernetes + Istio Ambient Mesh + Skaffold).

## ✅ O que já foi concluído
- [x] Provisionamento do cluster Minikube (`food-ddd`) com os limites definidos (4 CPUs, 8GB RAM).
- [x] Habilitação do addon `metrics-server` no Minikube.
- [x] Instalação do Istio CLI (`istioctl` v1.29.0).
- [x] Instalação do Istio no cluster usando o profile **Ambient**.
- [x] Deploy do `istio-ingressgateway` (via Helm interno do Istio).
- [x] Inclusão do namespace `default` na malha Ambient (`istio.io/dataplane-mode=ambient`).
- [x] Instalação da Stack de Observabilidade (Prometheus, Jaeger, Kiali, Grafana).
- [x] Criação do chart Helm genérico (`deploy/helm/ms-base`) focado em deploys enxutos sem sidecars, com suporte a variáveis baseadas em ConfigMap e labels para telemetria.
- [x] Criação/Uso do Kubernetes namespaces e manifestos para `catalog` e `shared-infra` (Brokers & Bancos In-Cluster).
- [x] Preparação básica do Sub-Chart / Values do Catalog.
- [x] Build and Deploy loop (`skaffold dev`) plenamente integrado ao cluster local.
- [x] Aplicação de Migrations no banco Kubernetes (`catalog-db`) via port-forwarding.
---

## ⏳ Próximos Passos (Backlog)

### 4. Configuração Avançada do Istio (L7 Waypoint Proxy & JWT)
*(Importante: Estas etapas afetam diretamente o código Go do Catalog)*
- [x] Adicionar um proxy L7 intermédio (Waypoint) no namespace `catalog` para habilitar as ricas policies de segurança do Istio.
- [x] Configurar o `RequestAuthentication` apontando para o Issuer/JWKS.
- [x] 🐞 **CORRIGIR:** O `AuthorizationPolicy` criado (no helm chart base) falhou com "Unsupported value: REQUIRE" (trocar para `ALLOW`).
- [x] Ajustar o código Go do Catalog (Middleware/Interceptors) para:
  - Extrair os headers (JWT) injetados pelo Istio;
  - Extrair informações dos Traces (`b3`) e passar adiante no contexto (`Context.Context`) para que eventos do Outbox retenham correlação visual no Jaeger.

### 5. Pendências Recém-Mapeadas (Domínio/Código)
- [x] (Domain) Refatorar rota `GET /restaurant/{id}/menu` (Implementar CQRS de leitura contornando o Repository do Agregado).
- [x] (Domain) Enriquecer todos os testes atuais mapeando as novas validações e errors explícitos (ex: NotFound).

### 6. Configurar Ambiente Local (Gradativo - Novos Serviços)
- Organizar a estrutura de `docker-compose.yaml` (Fase "fora" do mesh) para as etapas de inicialização dos novos módulos (ex: Ordering).
- À medida que os novos módulos amadurecem, integrá-los progressivamente à malha K8s de desenvolvimento adicionando-os no `skaffold.yaml` e criando seus referidos `values.yaml`.

---

**Nota:** O fluxo de inicialização original está documentado em `.agent/workflows/infra-setup.md`. Use-o como referência principal para comandos do Zsh.
