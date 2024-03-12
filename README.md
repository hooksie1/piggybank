# Piggy Bank

Piggy Bank is a secrets storage tool for applications that works with NATS. Secrets are stored encrypted in JetStream and can be retrieved as long as the requestor has access to the subject.

A decryption key is returned from the initialization phase. If this key is lost, all of the data is unrecoverable.

## Add KV bucket

Be sure to add the KV bucket to NATS: `nats kv add piggybank`

## Example Usage

1. Start piggybank `piggybank start`
2. Initialize the database `nats req piggybankdb.initialize ""`
3. Unlock the database with key sent from step 1 `nats req piggybankdb.unlock '{"database_key": "foobar"}'`
4. Add a secret for an application `nats req -H method:post piggybank.myapplication.registrySecret "somesecrettext"`
5. Retrieve a secret `nats req -H method:get piggybank.myapplication.registrySecret ""`
6. Lock the database `nats req piggybankdb.lock ""`
7. Try to retrieve the secret again `nats req -H method:get piggybank.myapplication.registrySecret ""`

## Permissions
Permissions are defined as normal NATS subject permissions. If you have access to a subject, then you can retrieve the secrets. This means the permissions can be as granular as desired.

## Config
Piggy Bank requires a config file. It uses Cue to read the configs, but the configs can also be in json or yaml format.

The Cue schema is in `cmd/schema.cue`.
