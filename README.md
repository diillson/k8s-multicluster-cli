# Multicluster CLI

## Visão Geral

O Multicluster CLI é uma ferramenta desenvolvida para gerenciar múltiplos clusters Kubernetes simultaneamente. Ela oferece uma gama de comandos para interagir com recursos Kubernetes em diferentes clusters, como listar pods, nodes, ConfigMaps, Secrets, Ingressos, e mais. A CLI foi construída com escalabilidade e performance em mente, utilizando concorrência, caching e design modular para lidar eficientemente com operações em grande escala.

## Funcionalidades

- **Gerenciamento de Múltiplos Clusters**: Execute comandos em múltiplos clusters Kubernetes de forma concorrente.
- **Gerenciamento de Pods**: Liste pods em vários clusters com opções de filtro.
- **Gerenciamento de Nodes**: Liste e exiba o status de nodes, incluindo papéis e condições.
- **ConfigMaps e Secrets**: Liste e recupere ConfigMaps e Secrets, com a possibilidade de exibir o conteúdo.
- **Gerenciamento de Ingressos**: Recupere e exiba recursos de Ingress em vários clusters.
- **Logs**: Obtenha logs de pods, com opções para logs específicos de containers e streaming em tempo real (modo follow).
- **Operações com Manifestos**: Aplique e delete manifestos Kubernetes em múltiplos clusters.
- **Configurações Personalizáveis**: Suporte para arquivos de configuração personalizados e caminhos de kubeconfig.

## Instalação

### Pré-requisitos

- Go 1.22 ou superior
- Clusters Kubernetes configurados e acessíveis via kubeconfig

### Compilação da CLI

Clone o repositório e navegue até o diretório do projeto:

```bash
git clone https://github.com/seu-repositorio/multicluster-cli.git
cd multicluster-cli
```

Compile a CLI:

```bash
go build -o multicluster
```

## Uso

### Sintaxe Geral

```bash
./multicluster-cli [comando] [flags]
```

### Comandos

- **get pods**: Liste pods em múltiplos clusters.

  ```bash
  ./multicluster-cli get pods --config /caminho/para/config.json --kubeconfig /caminho/para/kubeconfig --namespaces default --status Running --cluster eks-cluster-1
  ```

- **get nodes**: Liste nodes em múltiplos clusters.

  ```bash
  ./multicluster-cli get nodes --config /caminho/para/config.json --kubeconfig /caminho/para/kubeconfig --cluster eks-cluster-1
  ```

- **get configmaps**: Liste ConfigMaps em múltiplos clusters.

  ```bash
  ./multicluster-cli get configmaps --config /caminho/para/config.json --kubeconfig /caminho/para/kubeconfig --namespaces default --name meu-configmap --cluster eks-cluster-1
  ```

- **get secrets**: Liste Secrets em múltiplos clusters.

  ```bash
  ./multicluster-cli get secrets --config /caminho/para/config.json --kubeconfig /caminho/para/kubeconfig --namespaces default --name meu-secret --cluster eks-cluster-1
  ```

- **get ingress**: Liste Ingresses em múltiplos clusters.

  ```bash
  ./multicluster-cli get ingress --config /caminho/para/config.json --kubeconfig /caminho/para/kubeconfig --namespaces default --cluster eks-cluster-1
  ```

- **get logs**: Obtenha logs de um pod específico.

  ```bash
  ./multicluster-cli get logs --config /caminho/para/config.json --kubeconfig /caminho/para/kubeconfig --namespace default --pod meu-pod --container meu-container --cluster eks-cluster-1 --follow
  ```

### Flags

- **`--config, -c`**: Caminho para o arquivo de configuração dos clusters (padrão: `config.json` ou valor de `MC_CONFIG`).
- **`--kubeconfig, -k`**: Caminho para o arquivo kubeconfig (padrão: `~/.kube/config`).
- **`--namespaces, -n`**: Lista de namespaces separada por vírgulas para filtrar (padrão: todos os namespaces).
- **`--status, -s`**: Status para filtrar pods (ex.: Running, Pending).
- **`--cluster, -l`**: Nome do cluster (se vazio, o comando se aplica a todos os clusters).

para mais comandos e flags tem o **--help** e o para atribuição curta **-h**

### Configuração

A CLI utiliza um arquivo de configuração JSON para definir clusters e contextos:

```json
{
  "clusters": [
    {
      "name": "eks-cluster-1",
      "context": "arn:aws:eks:region:id-conta:cluster/eks-cluster-1"
    },
    {
      "name": "eks-cluster-2",
      "context": "arn:aws:eks:region:id-conta:cluster/eks-cluster-2"
    }
  ]
}
```

### Caching

A CLI faz cache de dados de configuração e contexto para melhorar a performance. O cache expira após 5 minutos por padrão.

### Tratamento de Erros

A CLI utiliza `logrus` para logging estruturado e fornece mensagens detalhadas de erro. Se uma operação falhar em um cluster específico, a CLI continua processando os demais clusters.

### Exemplo de ouputs:

### Exemplos de Output para Comandos Comuns da Multicluster CLI

#### Comando: `multicluster get nodes`
```bash
Cluster: dev-cluster
+---------------+----------+---------------+--------+------------+
|     Name      |  Status  |     Roles     |  Age   |  Version   |
+---------------+----------+---------------+--------+------------+
| node-01       | Ready    | master        | 15d    | v1.22.0    |
| node-02       | Ready    | worker        | 15d    | v1.22.0    |
| node-03       | NotReady | worker        | 14d    | v1.22.0    |
| node-04       | Ready    | worker        | 13d    | v1.22.0    |
+---------------+----------+---------------+--------+------------+

Cluster: prod-cluster
+---------------+----------+---------------+--------+------------+
|     Name      |  Status  |     Roles     |  Age   |  Version   |
+---------------+----------+---------------+--------+------------+
| node-01       | Ready    | master        | 15d    | v1.22.0    |
| node-02       | Ready    | worker        | 15d    | v1.22.0    |
| node-03       | NotReady | worker        | 14d    | v1.22.0    |
| node-04       | Ready    | worker        | 13d    | v1.22.0    |
+---------------+----------+---------------+--------+------------+
```

#### Comando: `multicluster get pods --namespaces default,kube-system`
```bash
Cluster: staging-cluster
+------------+-------------------------+-----------------+---------+-------------+
| Namespace  |        Pod Name          | Containers READY| Status  |   Node      |
+------------+-------------------------+-----------------+---------+-------------+
| default    | frontend-deployment-1234 | 3/3             | Running | node-01     |
| default    | backend-deployment-5678  | 2/2             | Running | node-02     |
| kube-system| coredns-78fcdcb99c-kmdjl | 1/1             | Running | node-03     |
+------------+-------------------------+-----------------+---------+-------------+

Cluster: prod-cluster
+------------+-------------------------+-----------------+---------+-------------+
| Namespace  |        Pod Name          | Containers READY| Status  |   Node      |
+------------+-------------------------+-----------------+---------+-------------+
| default    | frontend-deployment-1234 | 3/3             | Running | node-01     |
| default    | backend-deployment-5678  | 2/2             | Running | node-02     |
| kube-system| coredns-78fcdcb99c-kmdjl | 1/1             | Running | node-03     |
+------------+-------------------------+-----------------+---------+-------------+
```

#### Comando: `get configmaps --cluster prod-cluster`
```bash
Cluster: prod-cluster
+-------------+-------------------+-------------------------+
| Namespace   |    ConfigMap Name |           Data          |
+-------------+-------------------+-------------------------+
| default     | app-config         | key1: value1            |
| kube-system | coredns            | Corefile: .:53...       |
+-------------+-------------------+-------------------------+
```

#### Comando: `multicluster get secrets --namespace default`
```bash
Cluster: dev-cluster
+-------------+-------------------+-----------+---------------------------------+
| Namespace   |    Secret Name    |   Type    |            Data Keys            |
+-------------+-------------------+-----------+---------------------------------+
| default     | db-password       | Opaque    | password: 16 bytes              |
| default     | api-key           | Opaque    | key: 32 bytes                   |
+-------------+-------------------+-----------+---------------------------------+

Cluster: prod-cluster
+-------------+-------------------+-----------+---------------------------------+
| Namespace   |    Secret Name    |   Type    |            Data Keys            |
+-------------+-------------------+-----------+---------------------------------+
| default     | db-password       | Opaque    | password: 16 bytes              |
| default     | api-key           | Opaque    | key: 32 bytes                   |
+-------------+-------------------+-----------+---------------------------------+
```

#### Comando: `multicluster get logs --namespace default --pod frontend-deployment-1234`
```bash
Cluster: dev-cluster
--- Logs for pod frontend-deployment-1234 in namespace default ---
2024-08-26 10:30:01 Starting application...
2024-08-26 10:30:02 Connecting to database...
2024-08-26 10:30:03 Application started on port 8080.
```

Esses exemplos mostram como a Multicluster CLI fornece informações detalhadas e organizadas para facilitar o gerenciamento de recursos Kubernetes em múltiplos clusters.

### Contribuição

Contribuições são bem-vindas! Por favor, faça um fork do repositório, crie um branch de feature e envie um pull request.

## Licença

Este projeto é licenciado sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.