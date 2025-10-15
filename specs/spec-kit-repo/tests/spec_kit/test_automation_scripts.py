"""
Integration tests for Script Automation (REQ-SK-501 to REQ-SK-504)

Tests verify that all bash automation scripts exist, are executable, and properly tracked.
"""

import os
import pytest
import stat
from pathlib import Path


# CANARY: REQ=REQ-SK-501; FEATURE="FeatureCreationScript"; ASPECT=Automation; STATUS=TESTED; TEST=test_feature_creation_script_exists; OWNER=tests; UPDATED=2025-10-15
def test_feature_creation_script_exists():
    """Test that feature creation script exists and is executable."""
    script_path = Path(__file__).parent.parent.parent / "scripts" / "bash" / "create-new-feature.sh"
    assert script_path.exists(), f"Feature creation script not found at {script_path}"

    # Verify it's executable
    file_stat = script_path.stat()
    assert file_stat.st_mode & stat.S_IXUSR, "Script should be executable by user"

    # Verify it contains CANARY token
    content = script_path.read_text()
    assert "CANARY:" in content, "Script should contain CANARY token"
    assert "REQ-SK-501" in content, "Script should track REQ-SK-501"

    # Verify shebang
    assert content.startswith("#!/"), "Script should have shebang"


# CANARY: REQ=REQ-SK-502; FEATURE="PlanSetupScript"; ASPECT=Automation; STATUS=TESTED; TEST=test_plan_setup_script_exists; OWNER=tests; UPDATED=2025-10-15
def test_plan_setup_script_exists():
    """Test that plan setup script exists and is executable."""
    script_path = Path(__file__).parent.parent.parent / "scripts" / "bash" / "setup-plan.sh"
    assert script_path.exists(), f"Plan setup script not found at {script_path}"

    # Verify it's executable
    file_stat = script_path.stat()
    assert file_stat.st_mode & stat.S_IXUSR, "Script should be executable by user"

    # Verify it contains CANARY token
    content = script_path.read_text()
    assert "CANARY:" in content, "Script should contain CANARY token"
    assert "REQ-SK-502" in content, "Script should track REQ-SK-502"

    # Verify it references plan setup functionality
    assert "plan" in content.lower(), "Script should reference plan functionality"


# CANARY: REQ=REQ-SK-503; FEATURE="AgentContextUpdate"; ASPECT=Automation; STATUS=TESTED; TEST=test_agent_context_update_script_exists; OWNER=tests; UPDATED=2025-10-15
def test_agent_context_update_script_exists():
    """Test that agent context update script exists and is executable."""
    script_path = Path(__file__).parent.parent.parent / "scripts" / "bash" / "update-agent-context.sh"
    assert script_path.exists(), f"Agent context update script not found at {script_path}"

    # Verify it's executable
    file_stat = script_path.stat()
    assert file_stat.st_mode & stat.S_IXUSR, "Script should be executable by user"

    # Verify it contains CANARY token
    content = script_path.read_text()
    assert "CANARY:" in content, "Script should contain CANARY token"
    assert "REQ-SK-503" in content, "Script should track REQ-SK-503"

    # Verify it references agent context
    assert "agent" in content.lower() or "context" in content.lower(), \
        "Script should reference agent context functionality"


# CANARY: REQ=REQ-SK-504; FEATURE="PrerequisitesCheck"; ASPECT=Automation; STATUS=TESTED; TEST=test_prerequisites_check_script_exists; OWNER=tests; UPDATED=2025-10-15
def test_prerequisites_check_script_exists():
    """Test that prerequisites check script exists and is executable."""
    script_path = Path(__file__).parent.parent.parent / "scripts" / "bash" / "check-prerequisites.sh"
    assert script_path.exists(), f"Prerequisites check script not found at {script_path}"

    # Verify it's executable
    file_stat = script_path.stat()
    assert file_stat.st_mode & stat.S_IXUSR, "Script should be executable by user"

    # Verify it contains CANARY token
    content = script_path.read_text()
    assert "CANARY:" in content, "Script should contain CANARY token"
    assert "REQ-SK-504" in content, "Script should track REQ-SK-504"

    # Verify it references prerequisite checking
    assert "prerequisite" in content.lower() or "check" in content.lower() or "require" in content.lower(), \
        "Script should reference prerequisite checking"


def test_all_automation_scripts_tracked():
    """Meta-test: Verify all automation scripts exist and are tracked."""
    scripts_dir = Path(__file__).parent.parent.parent / "scripts" / "bash"
    assert scripts_dir.exists(), "Bash scripts directory should exist"

    expected_scripts = [
        "create-new-feature.sh",
        "setup-plan.sh",
        "update-agent-context.sh",
        "check-prerequisites.sh"
    ]

    for script_name in expected_scripts:
        script_path = scripts_dir / script_name
        assert script_path.exists(), f"Script {script_name} not found"

        content = script_path.read_text()
        assert "CANARY:" in content, f"Script {script_name} missing CANARY token"

        # Verify executable permission
        file_stat = script_path.stat()
        assert file_stat.st_mode & stat.S_IXUSR, f"Script {script_name} should be executable"


def test_scripts_follow_bash_best_practices():
    """Meta-test: Verify scripts follow bash best practices."""
    scripts_dir = Path(__file__).parent.parent.parent / "scripts" / "bash"

    for script_path in scripts_dir.glob("*.sh"):
        content = script_path.read_text()

        # Verify shebang
        assert content.startswith("#!/"), f"{script_path.name} should have shebang"

        # Verify CANARY token exists
        assert "CANARY:" in content, f"{script_path.name} should contain CANARY token"
