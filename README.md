# opwire-agent

Program `opwire-agent` is a simple command line wrapper. It receives a request from a REST API client or a message queue broker and transports this request to a command line process (that is developed by any programming language). 

<!-- TOC -->

- [Architecture](#architecture)
- [Configuration](#configuration)
  - [Location](#location)
  - [Structure](#structure)
- [Cmdline programs](#cmdline-programs)
  - [Interaction](#interaction)
  - [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

<!-- /TOC -->

## Architecture

![Architecture](https://raw.github.com/opwire/opwire-agent/master/docs/assets/images/arch.png)

## Configuration

### Location

Support configuration file (`opwire-agent.cfg` or `opwire-agent.conf`) from (priority in top-down order):

* `--config` command line argument,
* `OPWIRE_AGENT_CONFIG_DIR` environment variable,
* current binary directory (i.e the folder that contained `opwire-agent`),
* current working directory,
* XDG configuration directory (i.e `$HOME/.config/opwire-agent.conf`),
* Hidden configuration file in home directory (i.e `$HOME/.opwire-agent.conf`),
* `/etc` directory (i.e `/etc/opwire-agent.conf`).

### Structure

Configuration file description:

* `version`
* `agent`
  * `explanation`
    * `enabled`
    * `format`
  * `combine-stderr-stdout`
* `http-server`
  * `host`
  * `port`
  * `baseurl`
  * `read-timeout`
  * `write-timeout`
  * `concurrent-limit`
    * `enabled`
    * `total`
  * `single-flight`
    * `enabled`
    * `req-id`
    * `by-method`
    * `by-path`
    * `by-headers`
    * `by-queries`
    * `by-userip`
* `main-resource`
  * `enabled`
  * `pattern`
  * `default`
    * `command`
    * `timeout`
  * `methods`
    * `GET`
      * `command`
      * `timeout`
    * `POST`
      * `command`
      * `timeout`
    * `PATCH`
      * `command`
      * `timeout`
    * `PUT`
      * `command`
      * `timeout`
    * `DELETE`
      * `command`
      * `timeout`
  * `settings`
  * `settings-format`
* `resources`
  * `<NAME_OF_RESOURCE>`
    * `enabled`
    * `pattern`
    * `default`
      * `command`
      * `timeout`
    * `methods`
      * `GET`
        * `command`
        * `timeout`
      * `POST`
        * `command`
        * `timeout`
      * `PATCH`
        * `command`
        * `timeout`
      * `PUT`
        * `command`
        * `timeout`
      * `DELETE`
        * `command`
        * `timeout`
    * `settings`
    * `settings-format`
* `logging`
  * `enabled`
  * `format`
  * `level`
  * `output-paths`
  * `error-output-paths`

Configuration file in JSON pseudo-code:

```javascript
{
  "version": "<VERSION_OF_OPWIRE_AGENT>",
  "agent": {
    "explanation": {
      "enabled": true
    }
  },
  "main-resource": {
    "default": {
      "command": "<COMMAND LINE>",
      "timeout": 0 // no timeout by default
    },
    "methods": {
      "GET": {
        "command": "<COMMAND LINE FOR GET/LOAD/VIEW ACTION>"
      },
      "POST": {
        "command": "<COMMAND LINE FOR POST/INSERT/CREATE ACTION>"
      },
      "PUT": {
        "command": "<COMMAND LINE FOR PUT/REPLACE/UPDATE ACTION>"
      },
      "PATCH": {
        "command": "<COMMAND LINE FOR PATCH/MODIFY ACTION>"
      },
      "DELETE": {
        "command": "<COMMAND LINE FOR DELETE/REMOVE ACTION>"
      }
    },
    "settings": {
      "<YOUR_PARAM_1>": "<Text_Val_1",
      "<YOUR_PARAM_2>": "<Text_Val_2"
    },
    "settings-format": "json" // "json" or "flat"
  },
  "resources": {
    "<NAME_OF_RESOURCE>": {
      "default": {
        "command": "<COMMAND LINE>",
        "timeout": 30 // seconds
      }
    },
    // ...
  },
  "http-server": {
    "host": "<YOUR-BINDING-ADDR>",
    "port": 8888, // default: 17779
    "baseurl": "/run", // default: "/-"
    "read-timeout": "60s", // default: 30s
    "write-timeout": "90s" // default: 30s
  }
}
```

Example:

```javascript
{
  "version": "v1.0.8",
  "main-resource": {
    "default": {
      "command": "echo 'Hello opwire-agent'"
    }
  },
  "resources": {
    "products": {
      "pattern": "/api/v1/products",
      "default": {
        "command": "node product.js --action=list",
        "timeout": 5
      }
    },
    "product-create": {
      "pattern": "/api/v1/product",
      "methods": {
        "POST": {
          "command": "node product.js --action=create"
        }
      }
    },
    "product": {
      "pattern": "/api/v1/product/{productId}",
      "methods": {
        "GET": {
          "command": "node product.js --action=details"
        },
        "PUT": {
          "command": "node product.js --action=update"
        },
        "PATCH": {
          "command": "node product.js --action=change"
        },
        "DELETE": {
          "command": "node product.js --action=remove"
        }
      },
      "settings": {
        "MYSQL_URL": "mysql://localhost:3306",
        "MYSQL_USERNAME": "root",
        "MYSQL_PASSWORD": "root"
      },
      "settings-format": "json"
    }
  },
  "logging": {
    "format": "flat",
    "level": "debug",
    "output-paths": ["stdout"],
    "error-output-paths": ["stdout"]
  }
}
```

## Cmdline programs

### Interaction

Command line programs use 5 technical mechanism to communicate with outer service (i.e `opwire-agent`):

* Environment variables;
* Command arguments;
* Standard I/O: stdin, stdout, stderr;
* JSON encoding, decoding;
* Log messages (to log files);

### Examples

Opwire development team provides a collection of command line examples in several programming languges:

* [Command line example in Bash](https://github.com/opwire/sample-cmdline-bash)
* [Command line example in Java](https://github.com/opwire/sample-cmdline-java)
* [Command line example in Node.js](https://github.com/opwire/sample-cmdline-node)
* [Command line example in PHP](https://github.com/opwire/sample-cmdline-php)
* [Command line example in Python](https://github.com/opwire/sample-cmdline-python)
* [Command line example in Perl](https://github.com/opwire/sample-cmdline-perl)
* [Command line example in R](https://github.com/opwire/sample-cmdline-R)
* [Command line example in .NET](https://github.com/opwire/sample-cmdline-dotnet)

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b your-new-feature`)
3. Commit your changes (`git commit -am "Add some feature"`)
4. Push to the branch (`git push origin your-new-feature`)
5. Create new Pull Request

## License

MIT

See [LICENSE](LICENSE) to see the full text.
