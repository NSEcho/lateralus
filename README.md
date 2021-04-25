## What is Lateralus and why?

Lateralus is tool built to help with phishing campaigns. The reason I build this is to use it in conjugtion with modlishka. I can take _trackingParam_ from [modlishka](https://github.com/drk1wi/Modlishka), and use that to generate emails with lateralus.

## How does it work?
Steps to get it working:
* Add email template
* Add targets
* Choose mode(single url for all users, or each unique)
* Configure SMTP options
* Launch and wait

## Installation

```
$ git clone https://github.com/lateralusd/lateralus.git
$ cd lateralus/
$ go build
$ ./lateralus -help
Usage of ./lateralus:
  -attackerName string
    	Attacker name to use in template
  -config string
    	Config file to read parameters from
  -custom string
    	Custom words to include in template
  -delay int
    	delay between sending mails in seconds
  -from string
    	From field for an email. If not provided, will be the same as attackerName
  -generate
    	If set to true, parameter url needs to have <CHANGE> part
  -generateLength int
    	Length of variable part of url with maximum of 36 (default 8)
  -parseMdl string
    	Path to Modlishka control db file
  -priority string
    	priority to send email, can be low or high (default "low")
  -report string
    	Report name
  -signature string
    	path to signature .html file
  -singleUrl
    	Use the same URL for all targets (default true)
  -smtpConfig string
    	SMTP config file (default "conf/smtp.conf")
  -subject string
    	Subject that will be used for emails (default "Mail Subject")
  -targets string
    	File consisting of targets data (name, lastname, email, url)
  -template string
    	Email template from templates/ directory
  -templateName string
    	Email template name
  -url string
    	Single url to include in emails
```

## Setting up

### Creating template
The first step is to create the email template which you will be sending to your targets. Possible fields inside template are

* {{.Name}} - This will be substituted for target name from .csv file
* {{.URL}} - URL to include inside email
* {{.AttackerName}} - It says it all for itself

Example of template file can be found at `templates/sample.com`:
```
Greetings {{.Name}},

My resume is available at following url {{.URL}}

Best regards,
{{.AttackerName}}
```

### Creating targets
Targets needs to be in .csv format in format _Name_,_Email_ like so:
```
John,john.doe@example.com
Alan,alan.smith@example.com
```

### Choosing URL mode
If you choose to include URL inside your template, you have two options:
1. _Single URL_ mode
2. _Generate URL_ mode

In _Single URL_ mode, every user gets the same URL inside their body email. You activate this mode by providing `-singleUrl=true` flag to lateralus.

In _Generate URL_ mode, every user gets different URL. In order to use this mode, flag `-singleUrl` needs to be set to __false__, flag `-generate` needs to be set to __true__.  
Inside your URL you need to add `<CHANGE>` part. This is saying which part of the url you want to change. Generate URL mode will generate this part of `-generateLength` length.

For both of these modes, you need to pass `-url` parameter.

#### Example
Lets say we invoke literalus like 
```
lateralus \
-url https://www.example.com/?employeeID=<CHANGE> \
-singleUrl=false
-generate=true
-generateLength 10
```

URL that will be generated would look something like this:
```
https://www.example.com/?employeeID=ef4e7a4e-6
https://www.example.com/?employeeID=c5d568e2-4
https://www.example.com/?employeeID=727ef2df-2
```

### Configuring SMTP config file
Pass `-smtpConfig` flag with path to .json config file. It should look like:
```
{
"host": "smtp.gmail.com",
"port": 587,
"username": "username@gmail.com",
"password": "yourPassword",
"useSsl": false,
"useTls": true
}
```

## Example run
I will use two of my gmail accounts, one for sending, one for receiving. Targets file will have three targets with the same email and different name.

__NOTE: I do not own accounts which are given below, they are just there to show how it should look like__

File _templates/sample.com_:
```
Greetings {{.Name}},

My resume is available at following url {{.URL}}

Best regards,
{{.AttackerName}}
```

File _targets.csv_:
```
test,test@gmail.com
test1,test@gmail.com
test2,test@gmail.com
```

File _conf/smtp.conf_:
```
{
"host": "smtp.gmail.com",
"port": 587,
"username": "testuser@gmail.com",
"password": "testPassword",
"useSsl": false,
"useTls": true
}
```

File _conf/config.json_:
```
{
"singleUrl": false,
"config": "",
"template": "templates/sample.com",
"targets": "targets.csv",
"generateUrl": true,
"generateLength": 10,
"templateName": "",
"attackerName": "Attacker Himself",
"url": "https://www.example.com/?employeeID=<CHANGE>",
"custom": "",
"from": "Sample user",
"subject": "this is test"
}
```

You can either call lateralus with passing all the flags you need, or you can just pass `-config` parameter with path to your json configuration file.

Let's run it now.
```bash
$ lateralus -config conf/config.json
INFO[0000] Read 3 targets from targets.csv
INFO[26.04.2021 00:40:39] lateralus started
INFO[26.04.2021 00:40:39] Generating uuids for 3 users with uuid length: 36
INFO[26.04.2021 00:40:39] Sending mails
Sending mails: 3 / 3 [============================================================================================================================================>] 2 mail/s 100.00%INFO[26.04.2021 00:43:39] Report created at report_04-26-2021 00:39:43.txt
```

If we check inbox of user test@gmail.com, we can see that email has been sent.

![Mail](mailbox.png)

## Why lateralus as a name
I really love that album.
