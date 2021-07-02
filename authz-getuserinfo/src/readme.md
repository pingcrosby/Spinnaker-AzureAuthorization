## Tips

+ default port is 8080 - i have to override in the src dir because its in use
+ the config ends up as a secret in k8s (noting particularly secret in it)
+ watch the mappings as they define the fields spat out that your `~/.hal/default/profiles/gate-local.yml` will map to
+ NOTE that the k8s has its own config - which doesnt have a port override so on k8s its port `8080`
+ turn on trace for token tracing


Useful commands

To run..
```
export CONFIGPATH=$PWD/config.yaml
go run getuserinfo
```

Test docker after you have built it
```
docker run -e CONFIGPATH=/usr/opt/config.yaml -v $PWD:/usr/opt -p 8080:8080 authz-getuserinfo:latest
```
