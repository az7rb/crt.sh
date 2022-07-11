## CirrusGo
A fast tool to scan SAAS,PAAS vulnerability written in Go

SAAS App Support :

- salesforce
- contentful (next version)

**Note flag -o output not working**

**install :**

```
go install -v github.com/Ph33rr/CirrusGo/cmd/cirrusgo@latest
```
**Example :**

```
cirrusgo salesforce -u https://loclhost -gobj
```
**dump:**

```
cirrusgo salesforce -u https://localhost/ -f
```
**check Writable Objects:**

```
cirusgo salesforce -u https://localhost/ -cw
```

**Example Help:**

```
cirrusgo --help
```

 ```
   ______ _                           ______
  / ____/(_)_____ _____ __  __ _____ / ____/____
 / /    / // ___// ___// / / // ___// / __ / __ \
/ /___ / // /   / /   / /_/ /(__  )/ /_/ // /_/ /
\____//_//_/   /_/    \__,_//____/ \____/ \____/ v0.0.1

cirrusgo --help

-u, --url <URL>           Define single URL to fuzz
-l, --list		  Show App List
-c, --check               only check endpoint
-V, --version             Show current version
-h, --help                Display its help

[cirrusgo [app] [options] ..]
cirrusgo salesforce --help

-u, --url <URL>           Define single URL
-c, --check               only check endpoint
-lobj, --listobj          pull the object list.
-gobj --getobj            pull the object.
-obj --objects            set the object name. Default value is "User" object.
                           Juicy Objects: Case,Account,User,Contact,Document,Cont
                           entDocument,ContentVersion,ContentBody,CaseComment,Not
                           e,Employee,Attachment,EmailMessage,CaseExternalDocumen
                           t,Attachment,Lead,Name,EmailTemplate,EmailMessageRelation
-gre --getrecord          pull the Record id.
-re --recordid            set the recode id to dump the record
-cw --chkWritable         check all Writable objects
-f, --full                dump all pages of objects.
--dump
-H, --header <HEADER>     Pass custom header to target
-proxy, --proxy <URL>     Use proxy to fuzz

-o, --output <FILE>       File to save results

[flags payload]
[command: cirrusgo salesforce --payload options]
-payload, --payload      Generator payload for test manual Default "ObjectList"

GetItems                -obj set object
                         -page set page
                         -pages set pageSize
GetRecord 	        -re set recoder id 
WritableOBJ             -obj set object  
SearchObj               -obj set object 
                         -page set page
                         -pages set pageSize
AuraContext             -fwuid set UID 
                         -App set AppName
                         -markup set markup                        
ObjectList               no options
Dump                     no options		 
-h, --help               Display its help 

```

<img src="https://img.shields.io/badge/Open--Source--Summit-2022-blue.svg?logo=none" alt="" /></a>&nbsp;
[![made-with-Go](https://img.shields.io/badge/made%20with-Go-brightgreen.svg)](http://golang.org)
[![go-report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)
[![license](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/ph33rr/cirrusgo/issues)
[![godoc](https://img.shields.io/badge/godoc-reference-brightgreen.svg)](https://godoc.org/github.com/ph33rr/cirrusgo)
