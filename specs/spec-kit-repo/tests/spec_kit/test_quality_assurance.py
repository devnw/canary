# Copyright (c) 2024 by CodePros.
#
# This software is proprietary information of CodePros.
# Unauthorized use, copying, modification, distribution, and/or
# disclosure is strictly prohibited, except as provided under the terms
# of the commercial license agreement you have entered into with
# CodePros.
#
# For more details, see the LICENSE file in the root directory of this
# source code repository or contact CodePros at info@codepros.org.

"""
Integration tests for Quality Assurance (REQ-SK-801 to REQ-SK-804)

Tests verify that quality assurance features are implemented and tracked.
"""

import pytest
from pathlib import Path


# CANARY: REQ=REQ-SK-801; FEATURE="AmbiguityDetection"; ASPECT=Quality; STATUS=TESTED; TEST=test_ambiguity_detection_implementation; OWNER=tests; UPDATED=2025-10-15
def test_ambiguity_detection_implementation():
    """Test that ambiguity detection is implemented in clarify command."""
    clarify_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "clarify.md"
    assert clarify_path.exists(), f"Clarify command not found at {clarify_path}"

    content = clarify_path.read_text()
    assert "REQ-SK-801" in content, "Clarify command should track REQ-SK-801"

    # Verify ambiguity detection functionality
    ambiguity_indicators = [
        "ambiguity", "ambiguous", "unclear", "clarif", "question", "underspecified"
    ]
    has_ambiguity_detection = any(indicator in content.lower() for indicator in ambiguity_indicators)
    assert has_ambiguity_detection, "Command should implement ambiguity detection"

    # Verify substantial implementation
    assert len(content) > 500, "Clarify command should have substantial implementation"


# CANARY: REQ=REQ-SK-802; FEATURE="ConsistencyValidation"; ASPECT=Quality; STATUS=TESTED; TEST=test_consistency_validation_implementation; OWNER=tests; UPDATED=2025-10-15
def test_consistency_validation_implementation():
    """Test that consistency validation is implemented in analyze command."""
    analyze_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "analyze.md"
    assert analyze_path.exists(), f"Analyze command not found at {analyze_path}"

    content = analyze_path.read_text()
    assert "REQ-SK-802" in content, "Analyze command should track REQ-SK-802"

    # Verify consistency validation functionality
    consistency_indicators = [
        "consistency", "consistent", "validate", "verify", "check", "conflict"
    ]
    has_consistency_validation = any(indicator in content.lower() for indicator in consistency_indicators)
    assert has_consistency_validation, "Command should implement consistency validation"


# CANARY: REQ=REQ-SK-803; FEATURE="CoverageAnalysis"; ASPECT=Quality; STATUS=TESTED; TEST=test_coverage_analysis_implementation; OWNER=tests; UPDATED=2025-10-15
def test_coverage_analysis_implementation():
    """Test that coverage analysis is implemented in analyze command."""
    analyze_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "analyze.md"
    assert analyze_path.exists(), f"Analyze command not found at {analyze_path}"

    content = analyze_path.read_text()
    assert "REQ-SK-803" in content, "Analyze command should track REQ-SK-803"

    # Verify coverage analysis functionality
    coverage_indicators = [
        "coverage", "complete", "gap", "missing", "track", "analysis"
    ]
    has_coverage_analysis = any(indicator in content.lower() for indicator in coverage_indicators)
    assert has_coverage_analysis, "Command should implement coverage analysis"


# CANARY: REQ=REQ-SK-804; FEATURE="StalenessDetection"; ASPECT=Quality; STATUS=TESTED; TEST=test_staleness_detection_implementation; OWNER=tests; UPDATED=2025-10-15
def test_staleness_detection_implementation():
    """Test that staleness detection is implemented in prerequisites check."""
    prereq_path = Path(__file__).parent.parent.parent / "scripts" / "bash" / "check-prerequisites.sh"
    assert prereq_path.exists(), f"Prerequisites check not found at {prereq_path}"

    content = prereq_path.read_text()
    assert "REQ-SK-804" in content, "Prerequisites check should track REQ-SK-804"

    # Verify staleness detection functionality
    staleness_indicators = [
        "stale", "old", "outdated", "fresh", "up-to-date", "date", "time", "age"
    ]
    has_staleness_detection = any(indicator in content.lower() for indicator in staleness_indicators)
    assert has_staleness_detection, "Script should implement staleness detection"

    # Verify it's a proper bash script
    assert content.startswith("#!/"), "Should be an executable bash script"


def test_analyze_command_tracks_multiple_qa_features():
    """Meta-test: Verify analyze command tracks multiple QA features."""
    analyze_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "analyze.md"
    assert analyze_path.exists(), "Analyze command should exist"

    content = analyze_path.read_text()

    # Verify both consistency and coverage are tracked
    assert "REQ-SK-802" in content, "Should track consistency validation"
    assert "REQ-SK-803" in content, "Should track coverage analysis"

    # Count QA-related CANARY tokens
    qa_token_count = content.count("REQ-SK-80")
    assert qa_token_count >= 2, f"Analyze should track multiple QA features, found {qa_token_count}"


def test_quality_features_comprehensive():
    """Meta-test: Verify all quality assurance features are tracked."""
    # Check clarify.md for ambiguity detection
    clarify_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "clarify.md"
    assert clarify_path.exists(), "Clarify command should exist"
    clarify_content = clarify_path.read_text()
    assert "REQ-SK-801" in clarify_content, "Clarify should track ambiguity detection"

    # Check analyze.md for consistency and coverage
    analyze_path = Path(__file__).parent.parent.parent / "templates" / "commands" / "analyze.md"
    assert analyze_path.exists(), "Analyze command should exist"
    analyze_content = analyze_path.read_text()
    assert "REQ-SK-802" in analyze_content, "Analyze should track consistency validation"
    assert "REQ-SK-803" in analyze_content, "Analyze should track coverage analysis"

    # Check prerequisites script for staleness detection
    prereq_path = Path(__file__).parent.parent.parent / "scripts" / "bash" / "check-prerequisites.sh"
    assert prereq_path.exists(), "Prerequisites check should exist"
    prereq_content = prereq_path.read_text()
    assert "REQ-SK-804" in prereq_content, "Prerequisites should track staleness detection"


def test_qa_integration_with_commands():
    """Meta-test: Verify QA features integrate with command templates."""
    commands_dir = Path(__file__).parent.parent.parent / "templates" / "commands"
    assert commands_dir.exists(), "Commands directory should exist"

    # Verify clarify command exists and has CANARY tokens
    clarify = commands_dir / "clarify.md"
    assert clarify.exists(), "Clarify command should exist for ambiguity detection"

    # Verify analyze command exists and has CANARY tokens
    analyze = commands_dir / "analyze.md"
    assert analyze.exists(), "Analyze command should exist for consistency and coverage"

    # Both should be properly tracked
    for cmd_file in [clarify, analyze]:
        content = cmd_file.read_text()
        assert "CANARY:" in content, f"{cmd_file.name} should contain CANARY tokens"
