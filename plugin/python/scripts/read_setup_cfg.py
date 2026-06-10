import configparser, sys

config = configparser.ConfigParser()
config.read(sys.argv[1])
version = config.get("metadata", "version", fallback=None)

if version:
    print(version)
else:
    sys.exit(1)
