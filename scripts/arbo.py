import os

def generate_tree(root):
    tree = []
    for dirpath, dirnames, filenames in os.walk(root):
        level = dirpath.replace(root, '').count(os.sep)
        indent = ' ' * 4 * level
        tree.append(f"{indent}{os.path.basename(dirpath)}/")
        subindent = ' ' * 4 * (level + 1)
        for f in filenames:
            tree.append(f"{subindent}{f}")
    return "\n".join(tree)

def main():
    project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    tree = generate_tree(project_root)
    pour_claude_dir = os.path.join(project_root, "Pour_claude")
    os.makedirs(pour_claude_dir, exist_ok=True)
    with open(os.path.join(pour_claude_dir, "project_structure.md"), "w", encoding="utf-8") as f:
        f.write("# Project Structure\n\n")
        f.write("```\n")
        f.write(tree)
        f.write("```\n")

if __name__ == "__main__":
    main()