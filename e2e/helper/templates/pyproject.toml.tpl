[project]
name = "test-python-project"
version = "{{ .Version }}"
description = "Test Python project"
authors = [
    {name = "Test Author", email = "test@example.com"}
]
requires-python = ">=3.8"

[tool.hatch]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"