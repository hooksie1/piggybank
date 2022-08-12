#AppConfig: {
        urls:      string | *"nats://127.0.0.1:4222"
        creds?:    string
        user?:     string
        pass?:     string
        nkey?:     string
        tls_cert?: string
        tls_key?:  string
        tls_ca?:   string
}

Config: #AppConfig