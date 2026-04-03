---
description: Configuração de Infraestrutura K8s Local com Istio, Skaffold, Helm e Observabilidade
---

# 🚀 Workflow de Setup da Infraestrutura Kubernetes (Minikube + Istio)

Este workflow configura o ambiente local "parrudo" para desenvolvimento de microsserviços do projeto **Food Ordering DDD**. Ele provisiona um cluster Kubernetes (Minikube), instala o Service Mesh (Istio) com observabilidade (Kiali, Jaeger) e prepara o Setup de Deployment contínuo com Helm e Skaffold.

### Restrições:
- Instrua, e não saia executando comandos. Direcione o usuário através do chat com os comandos que precisam ser executados.

### ⚠️ Requisitos prévios:
- Docker instalado e rodando.
- [Minikube](https://minikube.sigs.k8s.io/docs/start/) instalado.
- [Helm](https://helm.sh/docs/intro/install/) instalado.
- [Skaffold](https://skaffold.dev/docs/install/) instalado.
- [Istioctl](https://istio.io/latest/docs/setup/getting-started/) instalado.

---

## Passo 1: Provisionar o Cluster Minikube

Criaremos um cluster com recursos suficientes para suportar o Istio e nosso stack de observabilidade.

// turbo
```bash
minikube start --cpus=4 --memory=8192 --disk-size=30g --driver=docker --profile=food-ddd
```

Ative os addons essenciais do Minikube (Ingress e Metrics Server):

// turbo
```bash
minikube addons enable ingress -p food-ddd
minikube addons enable metrics-server -p food-ddd
```

---

## Passo 2: Instalar o Istio (Service Mesh)

Vamos instalar o Istio com o profile "demo", que é ideal para ambientes locais pois possui tracing em alto volume e observabilidade pré-configuradas.

// turbo
```bash
istioctl install --set profile=demo -y --context=food-ddd
```

Agora, vamos habilitar a injeção automática de sidecars (Envoy) no namespace `default`. Todos os pods que fizermos deploy aqui entrarão automaticamente na malha.

// turbo
```bash
kubectl label namespace default istio-injection=enabled --overwrite --context=food-ddd
```

---

## Passo 3: Configurar Observabilidade (Jaeger, Kiali, Prometheus)

O Istio fornece manifestos de demonstração para instalar uma stack de observabilidade rapidamente.

Instalando complementos (Prometheus, Grafana, Jaeger, Kiali):

// turbo
```bash
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.29/samples/addons/prometheus.yaml --context=food-ddd
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.29/samples/addons/jaeger.yaml --context=food-ddd
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.29/samples/addons/kiali.yaml --context=food-ddd
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.29/samples/addons/grafana.yaml --context=food-ddd
```

> **Acesso aos Dashboards**:
> Para acessar o Kiali (Topologia visual), rode num terminal separado: `istioctl dashboard kiali --context=food-ddd`
> Para acessar o Jaeger (Traces): `istioctl dashboard jaeger --context=food-ddd`

---

## Passo 4: Criar o Template Helm Reutilizável

Vamos criar um Chart Helm genérico chamado `ms-base` (Microservice Base) dentro de `deploy/helm`. Todos os nossos serviços (Catalog, Ordering, etc.) reaproveitarão este template.

// turbo
```bash
mkdir -p deploy/helm
helm create deploy/helm/ms-base
```

**Nota**: Após a criação, limparemos o diretório e substituiremos os arquivos gerados (Deployment, Service, Gateway, VirtualService, AuthorizationPolicy) por templates que implementem JWT, ConfigMaps customizados e roteamento L7 do Istio. (Existem instruções detalhadas sobre isso nos próximos loops do agente).

---

## Passo 5: Configurar o Skaffold

O Skaffold vai monitorar nosso código Go, construir a imagem Docker dentro do Minikube e fazer o deploy via Helm, tudo com "Live Reload" automático (`skaffold dev`).

Crie o arquivo base do skaffold `skaffold.yaml` na raiz do projeto. Ele fará o build do Catalog e injetará no helm chart:

```yaml
# Arquivo: skaffold.yaml
apiVersion: skaffold/v4beta10
kind: Config
metadata:
  name: food-ordering-ddd
build:
  # Usa o ambiente docker DENTRO do minikube, evitando push pra registry externo
  local:
    push: false
  artifacts:
    # Módulo Catalog
    - image: catalog-api
      context: catalog
      docker:
        dockerfile: Dockerfile
deploy:
  helm:
    releases:
      - name: catalog-release
        chartPath: ./deploy/helm/ms-base
        valuesFiles:
          - ./catalog/values.yaml
        setValueTemplates:
          image.repository: "{{.IMAGE_REPO_catalog_api}}"
          image.tag: "{{.IMAGE_TAG_catalog_api}}"
```

---

## Passo 6: Integração Inicial (Estratégia Gradativa)

1. Após configurar a base acima, rodaremos o Catalog usando Skaffold.
2. Visitaremos o código fonte do Catalog (como *Middlewares*, Interceptores e DB) para capturar o contexto JWT / TraceID propagado pelo Istio.
3. Iniciaremos novos microsserviços com `docker-compose.yaml` locais, usando as abstrações criadas, e ao finalizá-los, faremos o plug deles no K8s adicionando-os ao `skaffold.yaml`.

---
**Dica**: Para rodar o setup inicial, use este workflow permitindo a execução automática das ferramentas (Turbo Mode). Para iniciar o desenvolvimento do Catalog após essa base pronta, basta rodar `skaffold dev` e codar!