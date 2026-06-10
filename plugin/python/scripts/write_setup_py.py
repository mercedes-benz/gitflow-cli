import ast, sys

path, new_version = sys.argv[1], sys.argv[2]

with open(path) as f:
    source = f.read()

if not source.strip():
    with open(path, "w") as f:
        f.write(f'from setuptools import setup\n\nsetup(\n    version="{new_version}",\n)\n')
    sys.exit(0)

tree = ast.parse(source)

for node in ast.walk(tree):
    if isinstance(node, ast.Call):
        for kw in node.keywords:
            if kw.arg == "version" and isinstance(kw.value, ast.Constant):
                line_idx = kw.value.lineno - 1
                lines = source.split("\n")
                line = lines[line_idx]
                start = kw.value.col_offset
                end = kw.value.end_col_offset
                quote = line[start]
                lines[line_idx] = line[:start] + quote + new_version + quote + line[end:]
                with open(path, "w") as f:
                    f.write("\n".join(lines))
                sys.exit(0)

sys.exit(1)
