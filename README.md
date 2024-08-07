# Piggy Bank

Piggy Bank is a secrets storage tool for applications that works with NATS. Secrets are stored encrypted in a JetStream KV and can be retrieved as long as the requestor has access to the subject.

> [!CAUTION]
> A decryption key is returned from the initialization phase. If this key is lost, all of the data is unrecoverable.

## Add KV bucket

Be sure to add the KV bucket to NATS: `nats kv add piggybank`

## Example Usage

1. Start piggybank `piggybank service start`
2. Initialize the database `piggybank client database initialize`
3. Unlock the database with key sent from step 1 `piggybank client database unlock --key foo`
4. Add a secret for an application `piggybank client secret add --id foo --value bar`
5. Retrieve a secret `piggybank client secret get --id foo`
6. Lock the database `piggybank client database lock`
7. Try to retrieve the secret again `piggybank client secret get --id foo`

## Permissions
Permissions are defined as normal NATS subject permissions. If you have access to a subject, then you can retrieve the secrets. This means the permissions can be as granular as desired. 

> [!IMPORTANT]
> Please ensure to set proper permissions for inbox responses. It is recommended to not use the default _INBOX prefix for responses and to set custom inbox prefixes. This prevents apps from listening to secrets sent on other apps inboxes.

## Client 

To send requests to Piggybank you can either send NATS requests directly or use the helper client. Here's an example assuming the key for the secret is myapp.somesecret:

```
nc, _ := nats.Connect(someURL)

client := service.Client{
	Conn: nc,
}

msg, _ := client.Get("myapp.somesecret")

fmt.Println(msg)
```

## NATS Connection

Piggybank supports multiple auth methods for NATS. 

1. Your current NATS context
2. A path to a credentials file
3. Env vars for the JWT and SEED
