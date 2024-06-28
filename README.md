#### Installation:

```bash
git clone https://github.com/ionutboangiu/rpgen.git
cd rpgen
go build .
```


#### Usage:

Running `./rpgen` is equivalent to running:

```bash
./rpgen \
-start_value=0 \
-count=3 \
-increment=102400 \
-unit=1024 \
-unit_name=MB \
-path= \
-tenant=cgrates.org \
-filter=*string:~*req.Account:1001 \
-subject=main_balance_subj \
-attr=ap_rating 
```

If path is empty it will output to stdout, otherwise to the specified path.

