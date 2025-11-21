# simple-search-engine
This project implements a simple search engine using Go and Elasticsearch

## 프로젝트 목표

- Go 언어 + Elasticsearch를 활용한 한국어 검색 엔진 구축
- Google/Naver 수준의 지능형 검색 기능 구현

## 기술 스택

- **언어**: Go 1.22
- **검색 엔진**: Elasticsearch 8.11 (Nori 한국어 분석기)
- **프레임워크**: Gin
- **인프라**: Docker, Docker Compose

## 실행

- Kibana: http://localhost:5601
- Elasticsearch: http://localhost:9200

## References
- [Go by Example](https://gobyexample.com/)
- [Elasticsearch 공식 문서](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Nori 한국어 분석기](https://www.elastic.co/guide/en/elasticsearch/plugins/current/analysis-nori.html)