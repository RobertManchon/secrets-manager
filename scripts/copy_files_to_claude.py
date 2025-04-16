import os
import shutil

def copy_files_to_claude(src_root, dest_dir):
    for dirpath, dirnames, filenames in os.walk(src_root):
        # Skip the Pour_claude directory
        if 'Pour_claude' in dirpath:
            continue
        for file in filenames:
            if file.endswith('.go') or file == 'go.mod':
                src_file = os.path.join(dirpath, file)
                dest_file = os.path.join(dest_dir, file)
                shutil.copy2(src_file, dest_file)
                print(f"Copied {src_file} to {dest_file}")

def main():
    project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    pour_claude_dir = "D:\\Pour_claude"
    os.makedirs(pour_claude_dir, exist_ok=True)
    copy_files_to_claude(project_root, pour_claude_dir)

if __name__ == "__main__":
    main()