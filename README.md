# opwire-agent

Program `opwire-agent` is a simple command line wrapper. It receives a request from a REST API client or a message queue broker and transports this request to a command line process (that is developed by any programming language). 

## Architecture

![Architecture](https://raw.github.com/opwire/opwire-agent/master/docs/assets/images/arch.png)

## Configuration

Support configuration file (`opwire-agent.cfg` or `opwire-agent.conf`) from (priority in top-down order):

* `--config` command line argument,
* `OPWIRE_AGENT_CONFIG_DIR` environment variable,
* current binary directory (i.e the folder that contained `opwire-agent`),
* current working directory,
* XDG configuration directory (i.e $HOME/.config/opwire-agent.conf),
* Hidden configuration file in home directory (i.e $HOME/.opwire-agent.conf),
* `/etc` directory (i.e /etc/opwire-agent.conf).

## Command line programs

Command line programs use 5 technical mechanism to communicate with outer service (i.e `opwire-agent`):

* Environment variables;
* Command arguments;
* Standard I/O: stdin, stdout, stderr;
* JSON encoding, decoding;
* Message logs (to log files);

Opwire development team provides a collection of command line examples in several programming languges:

* [Command line example in Java](https://github.com/opwire/sample-cmdline-java)
* [Command line example in Node.js](https://github.com/opwire/sample-cmdline-node)
* [Command line example in PHP](https://github.com/opwire/sample-cmdline-php)
* [Command line example in Python](https://github.com/opwire/sample-cmdline-python)
* [Command line example in Perl](https://github.com/opwire/sample-cmdline-perl)
* [Command line example in R](https://github.com/opwire/sample-cmdline-R)
* [Command line example in .NET](https://github.com/opwire/sample-cmdline-dotnet)

## License

MIT

See [LICENSE](LICENSE) to see the full text.
