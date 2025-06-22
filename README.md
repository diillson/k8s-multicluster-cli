• Paridade CLI com kubectl
• Multi-cluster, handlers universais, comandos core, plugins, completion, etc.
• Pronto para onboarding de times, CI/CD, SRE, DevOps, onboarding, auditoria, automação.
  
--------

# 🚀 Multicluster CLI — O "kubectl" universal para múltiplos clusters Kubernetes

O **Multicluster CLI** é uma linha de comando moderna, inspirada no kubectl, feita para operar **simultaneamente em múltiplos clusters**. Permite gerenciar, auditar, automatizar e operar clusters em larga escala, com experiência idêntica ao kubectl, porém multicluster de verdade.

---

## ✨ Principais Funcionalidades

- **Comandos kubectl universais:**  
- `get`, `describe`, `apply`, `delete`, `patch`, `logs`, `exec`, `scale`, `label`, `annotate`, `rollout` (status, history, undo, restart), `port-forward`, `events`, `config`, `version`, `can-i`, `drain`, `cordon`, `uncordon`, `diagnostics`...
- **Suporte a todos recursos:**  
- Recursos built-in, CRDs, subresources, plugins externos.
- **Multi-cluster simultâneo:**  
- Execute qualquer comando em TODOS ou grupo(s) de clusters, em paralelo, sem scripts.
- **Output agrupado por cluster**, em formato table, YAML, JSON.
- **Plugins estilo kubectl:**  
- Plugins (`multicluster-foo`) são auto-descobertos e executáveis via `plugin` ou fallback.
- **Completion automático:**  
- Completa comandos, flags, clusters, resources (bash, zsh, fish, powershell).
- **Diagnóstico/Self-Test:**  
- Checagem rápida do health/config dos clusters.
- **CLI modular, Go idiomático**, seguro e pronto para empresa de qualquer porte.

---

## 🛠️ Instalação

### Pré-requisitos:
- Go 1.22+ (para build/local)
- Kubernetes clusters acessíveis e credenciais kubeconfig/context corretamente configuradas

### Build local

```sh
git clone https://github.com/diillson/k8s-multicluster-cli.git
cd k8s-multicluster-cli
go build -o multicluster
```  
--------

## 🚦 Como usar

### Configuração (config.json)

Arquivo JSON apontando nome do cluster e o contexto do kubeconfig correspondente:

    {
      "clusters": [
        { "name": "dev", "context": "dev-context" },
        { "name": "prod", "context": "prod-context" },
        { "name": "eks-1", "context": "arn:aws:eks:us-east-1:xxxxxxx:cluster/eks-1" }
      ]
    }

* O nome é usado por todas as flags  --cluster .
* Contextos precisam existir no seu kubeconfig.
  
--------

# Exemplos de comandos

## Get, Describe, Delete, etc

#### Get pods em todos clusters
multicluster get pods -n default

### Describe deployment em todos clusters
multicluster describe deployment minha-app -n dev

### Delete service em todos clusters (flag --cluster limita)
multicluster delete svc meu-svc -n prod --cluster dev

# Apply, Patch, Rollout, Label, Annotate

#### Apply yaml em todos clusters
multicluster apply -f manifest.yaml

### Patch (edit) universal
multicluster patch deployment minha-app -n prod --patch '{"spec":{"replicas":3}}' --type strategic

### Restart rollout universal
multicluster patch deployment minha-app -n prod --patch '{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"2024-05-18T20:00:00Z"}}}}}' --type strategic

### Escalar deployment/statefulset
multicluster scale deployment minha-app --replicas=4 -n prod

### Label e annotate
multicluster label deployment minha-app env=prod -n prod
multicluster annotate pod meu-pod foo=bar -n prod

# Logs, Exec, Port-forward, Events

### Logs de todos pods de um Job
multicluster logs job meu-job -n batch-jobs

### Exec (apenas em um cluster se for interativo)
multicluster exec meu-pod -- bash -l -n prod --cluster dev

### Port-forward (apenas um cluster por vez)
multicluster port-forward --pod meu-pod --ports 8000:80 -n prod

#### Config, Plugins, Completion, Diagnóstico

### Config
multicluster config view
multicluster config get-contexts

### Detecção/execução de plugins
multicluster plugin
multicluster plugin foo arg1

### Completion
multicluster completion bash   # (siga instruções para habilitar)

### Diagnóstico self-test multicluster
multicluster diagnostics

### RBAC quick-check
multicluster can-i create pod -n prod

--------

## 🧰 Flags principais

*  --config-file, --cf  : Caminho do config.json multicluster (padrão: config.json)
*  --kubeconfig         : Caminho kubeconfig base do kubectl (default: ~/.kube/config)
*  --cluster            : Executa só no cluster nomeado
*  --namespace, -n      : Namespace alvo (para recursos namespaced)
*  --output, -o         : Formato de saída (table|json|yaml)
* Para mais, cheque  --help
  
--------

## 🧩 Plugins (estilo kubectl)

* Basta instalar um binário executável com o prefixo  multicluster-foo  no seu PATH.
* Ele será descoberto e usado via  multicluster plugin foo , ou  multicluster foo ...  (via fallback).
  
--------

## 🛡️ Segurança e Enterprise

* Fully Go idiomático, thread-safe, sem ciclos de import, pronto para paralelismo e automação massiva.
* Output sempre agrupado e separados por cluster para rastreio de CI/CD, SRE e troubleshooting.
  
--------

## 📌 Diagnóstico/self-test

    multicluster diagnostics

Diagnostica credenciais, versionamento, RBAC, disponibilidade dos clusters.
  
--------

## 📖 Exemplos completos

Exemplos extras e scripts de onboarding: Veja o diretório  /examples  ou o wiki/documentação do repositório.
  
--------

## ❓ Dúvidas/complementos

* O config.json é simples — pode ser gerado manualmente ou via script.
* Tudo segue “kubectl UX” por padrão para fácil automação e adoção por times com experiência em Kubernetes.
  
--------

## 📜 Licença

MIT
  
--------

## 🤝 Contribuição

Contribuições são muito bem-vindas!, crie issues, PRs ou discussões.
  
--------

Multicluster CLI — Porque cloud de verdade é multi-cluster, multi-time, e agnóstico.