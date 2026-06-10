import ast, sys

with open(sys.argv[1]) as f:
    tree = ast.parse(f.read())

for node in ast.walk(tree):
    if isinstance(node, ast.Call):
        for kw in node.keywords:
            if kw.arg == "version" and isinstance(kw.value, ast.Constant):
                print(kw.value.value)
                sys.exit(0)

sys.exit(1)
