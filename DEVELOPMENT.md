# Development

## Entwicklung auf einem Entwicklungsserver

Die Anwendung benötigt einen [Secure context](https://developer.mozilla.org/en-US/docs/Web/Security/Secure_Contexts), um alle Funktionen ausführen zu können. (Kamera, Barcode API, ...)

Um einen Browser auf dem Entwicklungs-Client verwenden zu können, verwende ich Port Forwarding über einen SSH Tunnel. Das folgende Beispiel zeigt, wie ich einen Tunnel zum Entwicklungsserver *development.local* herstelle und den Port *9080* an den Entwicklungsclient weiterleite. Der Port ist auf dem Entwicklungsclient dann als Port *8080* verfügbar. Die Anwendung kann im Browser mit der URL *http://localhost:8080* erreicht werden.

```bash
ssh -L 8080:localhost:9080 development.local
```
