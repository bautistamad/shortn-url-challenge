Acortador de URLs

## High Level Design

<img width="1047" alt="image" src="https://github.com/bautistamad/shorten-url-challenge/assets/75149705/a8c48725-5356-4003-bf16-3492d71ab8db">

Para la demo seleccione una base de datos relacional, en gran escala iria por una no relacional.
## Como ejecutar

Para ejecutar el servicio es necesario tener docker-compose instalado, una vez instalado ejecutar el siguiente comando.

```
docker compose up
```

## Endpoints

#### Create URL Endpoint:
Request
```
POST http://localhost/shorten
{
    "long_url": "http://example.com/long-url"
}
```
Response:
```json
{
    "short_url": "http://example.com/{key}"
}
```

#### Redirect
```
GET http:/localhost/{key}
```
#### Delete URL:
Request
```
DELETE http://localhost/url/{key}
```
Response
```
{
  "message": "URL deleted successfully",
  "deleted_url": "https:/example.com/long-url"
}
```

### Get Data Endpoint:
Request
```
GET http://localhost/url/{key}/stats
```
Response
```
{
  "longURL": "https://example.com/long-url",
  "shortUrl": "http://localhost/Ipp7tz",
  "id": "26d05ccb-27f8-458a-99d5-5513855b29d9",
  "accessCount": 0
}
```
