# Beautifull Monalisa

This project is simple app about register user using React as front end and Go as a backend.

## Getting Started

This project written at go1.12 and using [godep]("https://github.com/golang/dep") as dependency manager.
Pick one package of dep from release tab that suitable to your system (i pick v0.5.4 as latest today).
Before running, make sure mysql running and bind into tcp/ip instead of socket, and configure the app
to point to correct database ip, port, username, and password. Configuration stored at ./config.toml file, template provided.

#### The config file template
```
[server]
bind = "0.0.0.0"
port = 5987

[database]
hostname = "localhost"
port = 3306
username = ""
password = ""
database = ""
```

#### Compiling process
```
$ git clone https://github.com/septianw/beautifull-monalisa.git
$ cd beautifull-monalisa
$ mv config-template.toml config.toml
$ godep ensure
$ cd ui; npm run build; cd ../; esc -prefix $(pwd)/ui/build -o static.go $(pwd)/ui/build; go build;
$ ./beautifull-monalisa
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

