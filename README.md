# invoice-microservice

## Comment lancer et compiler le microservice

Pour compiler, aller à la racine du projet et utiliser la commande 
```powershell
go build
```
Cette commande produit un exécutable qu'il suffit de lancer pour que le microservice soit actif.

## Comment accéder au microservice

Ce microservice se lance sur localhost:8002 par défaut. Pour en changer la configuration, modifiez le fichier main.go à la ligne 29 :
```go
err := http.ListenAndServe(":<port>", accountService.MakeHTTPHandler(service, logger))
```

Pour tester le microservice nous conseillons l'outil [Postman](https://www.postman.com) et [la collection fournie avec le microservice](https://github.com/PP-Groupe-6/invoice-microservice/blob/master/Invoices.postman_collection.json).

La liste des Url est la suivante :
| URL                     | Méthode           | Param (JSON dans le body) | Retour               |
| ----------------------- |:-----------------:| :------------------------:| :-------------------:|
| localhost:8002/invoices/  | GET             | {"ClientID": "\<ID\>", "CreatedBy": \<bool\>}      |{"invoices": [{"id": "\<ID\>","amount": \<amount\>,"state": "\<state : string\>","expDate": "\<expDate\>","withClientId": "\<withClientId\>"}, ...]}|
| localhost:8002/invoices/   | POST     | {"uid" : "\<user id\>","emailClient" : "\<emailClient\>","amount" : \<amount\>,"expDate" :"\<expDate\>"}|{"created": \<bool\>}|
| localhost:8002/invoices/pay  | POST              | {"Iid": "\<invoice id\>"} |{"paid": \<bool\>} |
| localhost:8002/invoices/ | DELETE             | {"Iid": "\<invoice id\>"} |{}|
