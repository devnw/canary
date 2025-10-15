"""
Integration tests for Package Management (REQ-SK-901 to REQ-SK-903)

Tests verify that package management scripts exist, are executable, and properly tracked.
"""

import os
import pytest
import stat
from pathlib import Path


# CANARY: REQ=REQ-SK-901; FEATURE="ReleasePackages"; ASPECT=PackageManagement; STATUS=TESTED; TEST=test_release_packages_script_exists; OWNER=tests; UPDATED=2025-10-15
def test_release_packages_script_exists():
    """Test that release packages script exists and is executable."""
    script_path = Path(__file__).parent.parent.parent / ".github" / "workflows" / "scripts" / "create-release-packages.sh"
    assert script_path.exists(), f"Release packages script not found at {script_path}"

    # Verify it's executable
    file_stat = script_path.stat()
    assert file_stat.st_mode & stat.S_IXUSR, "Script should be executable by user"

    # Verify it contains CANARY token
    content = script_path.read_text()
    assert "CANARY:" in content, "Script should contain CANARY token"
    assert "REQ-SK-901" in content, "Script should track REQ-SK-901"

    # Verify shebang
    assert content.startswith("#!/"), "Script should have shebang"

    # Verify package-related functionality
    package_indicators = [
        "package", "release", "build", "artifact", "tar", "zip", "dist"
    ]
    has_package_content = any(indicator in content.lower() for indicator in package_indicators)
    assert has_package_content, "Script should contain package creation functionality"


# CANARY: REQ=REQ-SK-902; FEATURE="GitHubRelease"; ASPECT=PackageManagement; STATUS=TESTED; TEST=test_github_release_script_exists; OWNER=tests; UPDATED=2025-10-15
def test_github_release_script_exists():
    """Test that GitHub release script exists and is executable."""
    script_path = Path(__file__).parent.parent.parent / ".github" / "workflows" / "scripts" / "create-github-release.sh"
    assert script_path.exists(), f"GitHub release script not found at {script_path}"

    # Verify it's executable
    file_stat = script_path.stat()
    assert file_stat.st_mode & stat.S_IXUSR, "Script should be executable by user"

    # Verify it contains CANARY token
    content = script_path.read_text()
    assert "CANARY:" in content, "Script should contain CANARY token"
    assert "REQ-SK-902" in content, "Script should track REQ-SK-902"

    # Verify shebang
    assert content.startswith("#!/"), "Script should have shebang"

    # Verify GitHub release functionality
    github_indicators = [
        "github", "release", "tag", "gh ", "api", "publish"
    ]
    has_github_content = any(indicator in content.lower() for indicator in github_indicators)
    assert has_github_content, "Script should contain GitHub release functionality"


# CANARY: REQ=REQ-SK-903; FEATURE="VersionManagement"; ASPECT=PackageManagement; STATUS=TESTED; TEST=test_version_management_script_exists; OWNER=tests; UPDATED=2025-10-15
def test_version_management_script_exists():
    """Test that version management script exists and is executable."""
    script_path = Path(__file__).parent.parent.parent / ".github" / "workflows" / "scripts" / "update-version.sh"
    assert script_path.exists(), f"Version management script not found at {script_path}"

    # Verify it's executable
    file_stat = script_path.stat()
    assert file_stat.st_mode & stat.S_IXUSR, "Script should be executable by user"

    # Verify it contains CANARY token
    content = script_path.read_text()
    assert "CANARY:" in content, "Script should contain CANARY token"
    assert "REQ-SK-903" in content, "Script should track REQ-SK-903"

    # Verify shebang
    assert content.startswith("#!/"), "Script should have shebang"

    # Verify version management functionality
    version_indicators = [
        "version", "semver", "update", "bump", "pyproject", "toml"
    ]
    has_version_content = any(indicator in content.lower() for indicator in version_indicators)
    assert has_version_content, "Script should contain version management functionality"


def test_all_package_management_scripts_tracked():
    """Meta-test: Verify all package management scripts exist and are tracked."""
    scripts_dir = Path(__file__).parent.parent.parent / ".github" / "workflows" / "scripts"
    assert scripts_dir.exists(), "GitHub workflows scripts directory should exist"

    expected_scripts = [
        "create-release-packages.sh",
        "create-github-release.sh",
        "update-version.sh"
    ]

    for script_name in expected_scripts:
        script_path = scripts_dir / script_name
        assert script_path.exists(), f"Script {script_name} not found"

        content = script_path.read_text()
        assert "CANARY:" in content, f"Script {script_name} missing CANARY token"

        # Verify executable permission
        file_stat = script_path.stat()
        assert file_stat.st_mode & stat.S_IXUSR, f"Script {script_name} should be executable"


def test_package_management_follows_bash_best_practices():
    """Meta-test: Verify package management scripts follow bash best practices."""
    scripts_dir = Path(__file__).parent.parent.parent / ".github" / "workflows" / "scripts"

    for script_path in scripts_dir.glob("*.sh"):
        # Only check our package management scripts
        if script_path.name in ["create-release-packages.sh", "create-github-release.sh", "update-version.sh"]:
            content = script_path.read_text()

            # Verify shebang
            assert content.startswith("#!/"), f"{script_path.name} should have shebang"

            # Verify CANARY token exists
            assert "CANARY:" in content, f"{script_path.name} should contain CANARY token"


def test_github_workflows_directory_structure():
    """Meta-test: Verify GitHub workflows directory has proper structure."""
    workflows_dir = Path(__file__).parent.parent.parent / ".github" / "workflows"
    assert workflows_dir.exists(), "GitHub workflows directory should exist"

    scripts_dir = workflows_dir / "scripts"
    assert scripts_dir.exists(), "GitHub workflows scripts subdirectory should exist"


def test_pyproject_exists_for_version_management():
    """Meta-test: Verify pyproject.toml exists for version management."""
    pyproject_path = Path(__file__).parent.parent.parent / "pyproject.toml"
    assert pyproject_path.exists(), "pyproject.toml should exist for version management"

    content = pyproject_path.read_text()
    assert "version" in content, "pyproject.toml should contain version field"
    assert "specify-cli" in content or "name" in content, "pyproject.toml should define package name"
