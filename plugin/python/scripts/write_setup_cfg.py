import configparser, sys

path, version = sys.argv[1], sys.argv[2]

config = configparser.ConfigParser()
config.read(path)

if not config.has_section("metadata"):
    config.add_section("metadata")

config["metadata"]["version"] = version

with open(path, "w") as f:
    config.write(f)
