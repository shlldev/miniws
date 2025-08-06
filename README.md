# miniws
miniws (minimalist web server) is a very simple web server written in golang. its purpose is to be lightweight, easy to configure and easily expandable for personal use. 

## command line arguments
```
  -h  --help           Print help information 
  -p  --port           what port miniws will run on. Default: 8040
  -l  --logs-folder    the logs folder. Default: logs
  -c  --config-folder  the configurations folder. Default: config
  -w  --www-folder     the www folder where miniws will look for files to
                       serve. Default: .
```

## how to configure
in your config folder you will find `ipfilter.conf` and `useragentfilter.conf`

both files use the same format: specify `allow|deny` in the first line to tell miniws to treat the file as a whitelist or a blacklist, then specify one ip/user-agent per line. 

## logging

in your logging folder you will find `access.log` and `errors.log`

`access.log` utilizes the NCSA **[Combined Log Format](http://fileformats.archiveteam.org/wiki/Combined_Log_Format)**

`errors.log` is for golang errors
