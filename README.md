## crt.sh

This bash script makes it easy to quickly save and parse the output from https://crt.sh website.
 to be sent to tools like httpx!

Usage is pretty simple :

![alt text](https://raw.githubusercontent.com/az7rb/crt.sh/main/Screenshot/Screenshot_Help.png)

Step 1:
```
git clone https://github.com/az7rb/crt.sh.git && cd crt.sh/ && chmod +x crt.sh
```
Step 2:
```
./crt.sh -h
```
Example :
```
./crt.sh -d hackerone.com | httpx
```

This will write all of the enumerated subdomains to the specified output file and will be ready to be passed to other tools.

Don't forget to visit [BugBountyzip]([https://github.com/BugBountyzip)

Happy hunting!
