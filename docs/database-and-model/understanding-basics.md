---
description: Covers how to work with the database layer.
---

# Understanding Basics

With `appy` framework, you can decide if you wanna use 1 of the below to interact with the database:

* `app.DB(name string) record.DBer` which returns a DB handle with the specified name. 
* `app.Model(dest interface{}, opts ...record.ModelOption) record.Modeler` which returns a ORM model.



