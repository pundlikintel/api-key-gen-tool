# API key generator tool 

#### How to use:

### Build
```bash
    go build -o api-key-gen .\src
```    
 
### Run

Make required change in [properties.toml](properties.toml) file

```bash
    .\api-key-gen
```

### Cleanup
#### Cleanup all records
```bash
    .\api-key-gen -cleanup all
```
#### Cleanup given number of records
.\api-key-gen -cleanup <number of record to clean>

```bash
    .\api-key-gen -cleanup 5
    
```

`Note`: Cleanup will use email_domain parameter from properties.toml file to delete all the tenants with that domain.