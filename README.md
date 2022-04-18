## crt.sh

This bash script makes it easy to quickly save and parse the output from https://crt.sh website.
 to be sent to tools like httpx!

Usage is pretty simple.

![alt text](https://i.ibb.co/sVbzXPc/Screenshot.png)

Step 1:
```
git clone https://github.com/az7rb/crt.sh.git
```
Step 2:
```
cd crt.sh/
```
![alt text](https://i.ibb.co/qsYkdqb/apikey.png)

Step 3:
```
chmod +x crt.sh
```
Step 4:
```
./crt.sh [domain.com]
```

This will write all of the enumerated subdomains to the specified output file and will be ready to be passed to other tools.

Happy hunting!
