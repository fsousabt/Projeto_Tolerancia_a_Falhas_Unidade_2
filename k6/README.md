# Testes de Carga com K6

## Pré-requisitos 

Instale o [K6](https://grafana.com/docs/k6/latest/) para executar os testes

## Como rodar os testes

1- Rode os microsserviços em ambiente local seguindo os passos do README.md na pasta raiz

2- Execute os testes com o k6, passando a variavel ft (true ou false) e o arquivo de teste:

```bash
k6 run -e ft=[true/false] file
```

Exemplo:

```bash
k6 run -e ft=true load-test.js
```
