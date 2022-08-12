# Piggy Bank

Piggy Bank is a secrets storage tool for applications that works with NATS. Secrets are stored encrypted in JetStream and can be retrieved as long as the requestor has access to the subject.

A decryption key is returned from the initialization phase. If this key is lost, all of the data is unrecoverable.

## Example Usage

1. Initialize the database `nats req piggybank.database.initialize ""`
2. Unlock the database with key sent from step 1 `nats req piggybank.database.unlock '{"database_key": "foobar"}'`
3. Add a secret for an application `nats req -H method:post piggybank.myapplication.registrySecret "somesecrettext"`
4. Retrieve a secret `nats req -H method:get piggybank.myapplication.registrySecret`
5. Lock the database `nats req piggybank.database.lock ""`
6. Try to retrieve the secret again `nats req -H method:get piggybank.myapplication.registrySecret`

## Permissions
Permissions are defined as normal NATS subject permissions. If you have access to a subject, then you can retrieve the secrets. This means the permissions can be as granular as desired.

## Config
Piggy Bank requires a config file. It uses Cue to read the configs, but the configs can also be in json or yaml format.

The Cue schema is in `cmd/schema.cue`.