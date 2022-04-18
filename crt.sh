#!/bin/bash

echo "
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      ..| search crt.sh v1.0 |..     |
+   site : crt.sh Certificate Search  +
|            Twitter: az7rb           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	"
	 
	search() {
	requestsearch="$(curl -s "https://crt.sh?q=%.$domain&output=json")"
		 
			 echo $requestsearch > search.txt
			 cat search.txt | jq ".[].common_name,.[].name_value"| cut -d'"' -f2 | sed 's/\\n/\n/g' | sed 's/\*.//g' | sort | uniq > output/$domain.txt
			 rm search.txt
			 echo ""
			 cat output/$domain.txt
			 echo ""
			 echo -e "\e[32m[+]\e[0m Total Save will be \e[31m"$(cat output/$domain.txt | wc -l)"\e[0m Domain only"
			 echo -e "\e[32m[+]\e[0m Output saved in output/$domain.txt"
}	 

if [ -z $1 ]
        then
                echo "USAGE: $0 [domain.com] "
                exit
        else
                domain=$1
fi

if [ -z $2 ]
then
search
fi