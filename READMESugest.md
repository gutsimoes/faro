# Faro — escopo e roteiro para apresentação

## Qual é o projeto?

O Faro é um motor de busca por similaridade vetorial desenvolvido em Go. Ele armazena vetores associados a identificadores e encontra os `k` vetores mais próximos de uma consulta.

Um vetor é uma lista de números que representa características de um item. Por exemplo, a distância entre `[0, 0]` e `[3, 4]` é 5 pela métrica euclidiana. A busca top-k calcula quais itens estão mais próximos da consulta e devolve seus IDs e distâncias.

O objetivo da disciplina é explorar as principais características de Go em um projeto concreto: organização em pacotes, tipos, tratamento de erros, concorrência, sincronização, testes e ferramentas de desempenho. Cada recurso deve resolver uma necessidade real do projeto.

## Para que o projeto é adequado?

O Faro é adequado para estudar algoritmos de busca vetorial e comparar suas escolhas de implementação: correção, latência, uso de memória e qualidade dos resultados. Também permite investigar quando dividir uma busca entre trabalhadores melhora o desempenho.

Motores desse tipo podem ser usados para encontrar itens semelhantes em sistemas de recomendação e em buscas sobre representações numéricas de textos ou imagens. Esses usos dependem de vetores gerados por outro componente e de uma métrica apropriada aos dados. O Faro atual usa somente vetores numéricos sintéticos; não gera embeddings nem realiza busca semântica de textos.

É um projeto acadêmico e experimental. Não está pronto para uso como banco de dados de produção.

## Escopo

O núcleo do projeto é o motor de similaridade:

- Armazenar vetores e IDs em memória.
- Calcular distâncias e executar busca exata top-k.
- Implementar uma versão paralela da busca exata e compará-la com a sequencial.
- Implementar HNSW para busca aproximada e comparar seus resultados com a referência exata.
- Medir latência, alocações e, na busca aproximada, recall@k.

API HTTP e persistência por snapshots são funcionalidades de apoio planejadas. Sharding, replicação, autenticação e arquitetura distribuída ficam fora do escopo.

## O que já fizemos

| Entrega | Implementação atual |
|---|---|
| Coleção em memória | `Collection` mantém vetores de dimensão fixa associados a IDs únicos |
| Representação dos dados | Structs, `map[string][]float64` e slices |
| Inserção | Copia o vetor recebido para evitar alterações externas nos dados armazenados |
| Distância | L2, também chamada de distância euclidiana |
| Busca exata | Percorre todos os vetores, calcula as distâncias, ordena e devolve até k resultados |
| Desempate | Ordenação por ID quando as distâncias são iguais |
| Validações | Dimensões, coordenadas finitas, IDs não vazios e únicos, k positivo |
| Casos especiais | Coleção vazia retorna resultado vazio; k maior que a coleção retorna todos os itens |
| Demonstração | Programa no terminal com quatro vetores e uma consulta |
| Testes | Distâncias conhecidas, ordenação, desempates, entradas inválidas e proteção da cópia inserida |
| Benchmark | Busca sequencial com 1.000 e 10.000 vetores, em 16 e 128 dimensões |
| CI | Workflow configurado para formatação, análise estática, compilação e testes |

A busca atual custa O(N × D + N log N) em tempo, pois calcula distâncias e ordena todos os resultados, e O(N) em memória auxiliar. N é a quantidade de vetores e D é o número de dimensões.

Arquivos para estudar antes da apresentação:

- [main.go](main.go): monta os dados e executa a demonstração.
- [vector/collection.go](vector/collection.go): coleção, validações, distância e busca.
- [vector/collection_test.go](vector/collection_test.go): testes e benchmark.
- [.github/workflows/ci.yml](.github/workflows/ci.yml): verificações automatizadas configuradas.

## Por que Go é uma boa escolha?

Go oferece compilação para código nativo, tipagem estática e uma biblioteca padrão que cobre testes, medição de desempenho, HTTP e JSON. Isso permite desenvolver o núcleo atual sem dependências externas.

Goroutines e channels são especialmente relevantes para o projeto porque os cálculos das distâncias entre a consulta e cada vetor são independentes. A linguagem permite organizar esses cálculos em tarefas concorrentes e coordenar seus resultados. O ganho de desempenho precisa ser demonstrado por medições.

| Recurso de Go | Aplicação no Faro | Situação |
|---|---|---|
| Pacotes e métodos | Separação do núcleo `vector` e da demonstração | Implementado |
| Structs, slices e maps | Coleção, coordenadas e resultados | Implementado |
| Retorno explícito de erros | Validação de dados e propagação de falhas | Implementado |
| Biblioteca padrão | Cálculo numérico, ordenação, testes e benchmark | Implementado |
| Goroutines | Trabalhadores para calcular distâncias | Planejado |
| Channels | Comunicação dos resultados dos trabalhadores | Planejado |
| `context.Context` | Cancelamento e timeout de buscas | Planejado |
| Interfaces pequenas | Contrato comum para os mecanismos de busca exata e HNSW | Planejado |
| Sincronização com `sync` | Coordenação dos trabalhadores e proteção de dados compartilhados | Planejado |
| `net/http` e JSON | API para inserir vetores e consultar vizinhos | Planejado |
| Detector de corrida | Verificação de acessos concorrentes durante os testes | Previsto para a etapa concorrente |

## Onde tem paralelismo?

**Ainda não existe paralelismo explícito na busca implementada.** O método `Search` calcula uma distância por vez, em um laço sequencial. Não usamos goroutines ou channels no núcleo atual, e a coleção não possui proteção para inserções e buscas simultâneas.

O ponto principal para introduzir paralelismo é o cálculo das distâncias da busca exata:

1. Dividir os vetores em grupos, com uma quantidade limitada de trabalhadores.
2. Iniciar uma goroutine por trabalhador.
3. Cada trabalhador calcular as distâncias de seu grupo e selecionar seus melhores candidatos.
4. Enviar os candidatos por um channel.
5. Combinar os candidatos e selecionar o top-k global, preservando o mesmo desempate por ID da busca sequencial.

Exemplo com 10.000 vetores e quatro trabalhadores:

```text
                         Consulta
                            |
                 Divisão dos 10.000 vetores
                            |
        +-------------+-------------+-------------+
        |             |             |             |
   Goroutine 1   Goroutine 2   Goroutine 3   Goroutine 4
   2.500 vetores 2.500 vetores 2.500 vetores 2.500 vetores
        |             |             |             |
        +-------------+-------------+-------------+
                            |
                  Candidatos via channel
                            |
                    Combinação do top-k
```

Esse desenho é uma proposta para a próxima etapa, não a arquitetura já implementada.

Concorrência significa organizar tarefas que podem progredir de forma independente. Paralelismo significa executar tarefas ao mesmo tempo. Goroutines permitem concorrência; o paralelismo depende dos núcleos disponíveis e da configuração de execução do Go, incluindo `GOMAXPROCS`.

Criar trabalhadores, comunicar resultados e combinar candidatos tem custo. Por isso, a versão paralela pode ser mais lenta para coleções pequenas. Os benchmarks devem comparar a versão sequencial e a paralela com os mesmos dados, consultas e k, variando dimensões, tamanho da coleção e número de trabalhadores.

Antes de permitir inserções simultâneas, será necessário definir como proteger a coleção ou trabalhar sobre uma visão estável dos dados. Channels, por si só, não tornam o map compartilhado seguro. O cancelamento também deve permitir que os trabalhadores terminem sem deixar goroutines bloqueadas.

## O que ainda vamos fazer

### Próxima prioridade: concorrência e paralelismo

- Manter a busca sequencial como referência de correção.
- Implementar busca paralela com goroutines e channels.
- Definir a estratégia de acesso seguro aos dados.
- Adicionar cancelamento com `context.Context`.
- Testar se as versões sequencial e paralela retornam os mesmos resultados, incluindo empates.
- Executar testes com `go test -race ./...` em ambiente compatível.
- Comparar latência e alocações; calcular o ganho como tempo sequencial dividido pelo tempo paralelo.

### Evolução do motor

- Adicionar cosseno e produto escalar, documentando o critério de ordenação de cada métrica.
- Implementar HNSW, um índice baseado em um grafo com múltiplas camadas para busca aproximada.
- Medir tempo de construção do índice, latência e recall@k.

A busca exata examina todos os vetores e fornece a referência dos vizinhos corretos. O HNSW procura reduzir a quantidade de candidatos examinados, podendo perder alguns desses vizinhos. Recall@k mede a proporção dos vizinhos de referência recuperados pela busca aproximada. Os experimentos devem definir como tratar empates na distância de corte.

### Funcionalidades de apoio

- API HTTP com JSON, validação e timeout.
- Persistência por snapshot para salvar e restaurar vetores e IDs.
- Documentação com resultados, configuração dos experimentos e limitações.

Para a disciplina, a prioridade deve ser demonstrar os recursos de Go e avaliar seus efeitos. HNSW, HTTP e persistência devem ser planejados conforme o tempo disponível, evitando comprometer a correção do núcleo.

## Como demonstrar o que existe hoje

Com Go 1.23 ou mais recente instalado, na raiz do projeto:

```sh
go run .
go test ./...
go vet ./...
go build ./...
go test ./vector -run '^$' -bench BenchmarkSearch -benchmem
```

Saída esperada da demonstração:

```text
Faro 0.0-dev — exact L2 search
Vectors: 4 | Query: [0 0] | k: 3
1. origin distance=0.0000
2. near   distance=1.4142
3. middle distance=5.0000
```

Os benchmarks usam semente aleatória fixa e k=10. A geração e a inserção dos dados ficam fora do tempo medido. `ns/op` representa o tempo por consulta; `B/op` e `allocs/op` representam as alocações por consulta. Os números variam conforme a máquina e devem ser comparados em condições equivalentes.

## Resposta curta para o professor

> Nosso projeto é o Faro, um motor de busca por similaridade vetorial. Escolhemos Go para explorar tipos, pacotes, tratamento de erros, concorrência e ferramentas de teste e desempenho. Já implementamos a coleção em memória, a distância euclidiana, a busca exata sequencial, testes e benchmarks. O paralelismo será aplicado ao cálculo das distâncias, dividindo os vetores entre goroutines e reunindo os candidatos por channels. Vamos verificar a equivalência dos resultados e medir quando essa divisão melhora o desempenho. Depois, pretendemos comparar a busca exata com HNSW e adicionar HTTP e persistência.

## O que precisamos saber explicar na apresentação final

- O que um vetor representa e o que significa proximidade pela métrica escolhida.
- Como uma consulta percorre a coleção e produz o top-k.
- Por que a busca exata serve de referência para avaliar HNSW.
- Como o trabalho foi dividido entre goroutines e como os resultados são combinados.
- Como tratamos erros, cancelamento e acesso a dados compartilhados.
- Quais medições sustentam as conclusões sobre desempenho.
- Quais funcionalidades realmente foram concluídas e quais limitações permanecem.

Para cada recurso implementado, devemos conseguir responder: qual problema ele resolve, por que foi escolhido, onde aparece no código e qual teste ou experimento demonstra seu comportamento.
