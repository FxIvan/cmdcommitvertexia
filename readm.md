Agregar commitgen a la variable de entorno PATH
```
go build -o commitgen main.go
sudo mv commitgen /usr/local/bin/
commitgen <ticket> <descripcion>
```

```
commitgen <ticket> <descripcion>
```
Ejemplo de como ejecutar:
commitgen 257 "Sacando texto de mas en aviso de registro"