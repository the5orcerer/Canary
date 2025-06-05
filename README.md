````markdown
# 🐦 Canary - OAST Canary Token Injector for Bug Bounty Hunting

Canary is a blazing-fast, multithreaded Go tool designed to inject OAST-powered canary tokens into URLs or raw HTTP requests.  
Perfect for bug bounty hunters to detect SSRF, XSS, and other data-leaking vulnerabilities silently.

---

## 🚀 Features

- ⚡ High-speed multithreaded injection (Go powered)
- 🔗 Automatic canary token generation (timestamp, custom string, or numeric range)
- 🧬 Support for URL and raw HTTP request modes
- 🛠️ Parameter targeting, prefix/suffix support, and original value preservation
- 🧾 Logging of changed lines and flexible output options
- 🔍 Verbose and quiet modes for your workflow

---

## 📦 Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/yourusername/canary.git
cd canary
go build -o canary
````

---

## 🧑‍💻 Usage

```bash
cat urls.txt | ./canary -oa .oast.me -c timestamp -o output.txt
```

Or with raw HTTP request input:

```bash
./canary -f raw_requests.txt -r -oa .oast.me -c mytoken -a -k -l log.txt
```

---

## 🧰 Options

| Flag                | Description                                                                      |
| ------------------- | -------------------------------------------------------------------------------- |
| `-oa`, `--oast`     | OAST domain to use (e.g. `.oast.me`)                                             |
| `-c`, `--canary`    | Canary mode: `timestamp`, custom string (e.g. `mytoken`), or range (e.g. `1-10`) |
| `-o`, `--output`    | Output file path (default: `requests.dreq`)                                      |
| `-s`, `--suffix`    | Add suffix to token                                                              |
| `-p`, `--prefix`    | Add prefix to token                                                              |
| `-d`, `--delimiter` | Delimiter between token and OAST (e.g. `-` → `timestamp-oast.me`)                |
| `-k`, `--keep`      | Keep original param value and append token                                       |
| `-a`, `--all`       | Replace all matching parameters                                                  |
| `-t`, `--target`    | Comma-separated list of target parameters                                        |
| `-f`, `--file`      | Read input from file                                                             |
| `-r`, `--raw`       | Enable raw HTTP request mode                                                     |
| `-l`, `--log`       | Log replaced lines to a file                                                     |
| `-v`, `--verbose`   | Verbose mode                                                                     |
| `-q`, `--quiet`     | Quiet mode                                                                       |
| `-h`, `--help`      | Show help menu                                                                   |

---

## 📌 Example Scenarios

Inject a timestamp-based token into all parameters:

```bash
cat urls.txt | canary -oa .oast.site -c timestamp -o output.txt
```

Use a static token with a suffix and delimiter:

```bash
cat list.txt | canary -oa .oast.live -c mycanary -s ".track" -d "-" -o traced.txt
```

Use range-based canary tokens on only specific parameters:

```bash
canary -f data.txt -c 1-100 -t id,user -a -k -oa .oast.me -l changed.log
```

---

## 📄 Output

* Injected lines go to your specified output file.
* Logs (before → after) can optionally be saved using `-l`.
* No input? A friendly message and exit — no stack traces here.

---

## 🤝 Contributing

Found a bug or want to improve it? PRs are welcome!
Help evolve this tool for the bounty hunting community.

---

## 🧙‍♂️ Author

Crafted with ❤️ by [@rootplinix](https://github.com/rootplinix)
