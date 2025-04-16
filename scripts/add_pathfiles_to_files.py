import os

def add_path_to_go_files(src_root):
    for dirpath, dirnames, filenames in os.walk(src_root):
        # Skip the Pour_claude directory
        if 'Pour_claude' in dirpath:
            continue
        for file in filenames:
            if file.endswith('.go'):
                file_path = os.path.join(dirpath, file)
                with open(file_path, 'r+', encoding='utf-8') as f:
                    content = f.read()
                    f.seek(0, 0)
                    f.write(f"// filepath: {file_path}\n{content}")
                print(f"Added path to {file_path}")

def main():
    project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    add_path_to_go_files(project_root)

if __name__ == "__main__":
    main()