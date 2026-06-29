#!/usr/bin/env python3
"""
AGI Cognitive Benchmark Experiment Runner.

Executes benchmark tasks under controlled architectural conditions (Groups A/B/C),
records metrics, and produces the experiment dataset.

Usage:
  python3 experiment/runner.py --group A --domain CODE
  python3 experiment/runner.py --group A --task-ids CODE-001,CODE-002
  python3 experiment/runner.py --group B --all
"""

import json
import os
import sys
import time
import random
import argparse
from datetime import datetime, timezone
from pathlib import Path
from collections import defaultdict

EXPERIMENT_DIR = Path(__file__).parent
RESULTS_DIR = EXPERIMENT_DIR / "results"
TASKS_FILE = EXPERIMENT_DIR / "benchmark_tasks.json"
CONFIG_FILE = EXPERIMENT_DIR / "config.json"


class ExperimentRunner:
    """Orchestrates benchmark task execution and metrics collection."""

    def __init__(self, group: str):
        self.group = group.upper()
        if self.group not in ("A", "B", "C"):
            raise ValueError(f"Invalid group: {self.group}. Must be A, B, or C.")

        self.config = self._load_json(CONFIG_FILE)
        self.group_config = self.config["groups"][self.group]
        self.tasks_data = self._load_json(TASKS_FILE)
        self.all_tasks = self.tasks_data["tasks"]
        self.metrics_schema = self.config["metrics"]

        # Ensure results directory exists
        RESULTS_DIR.mkdir(parents=True, exist_ok=True)

        # Results file per group
        self.results_file = RESULTS_DIR / f"results_group_{self.group}.jsonl"
        self.summary_file = RESULTS_DIR / f"summary_group_{self.group}.json"

    @staticmethod
    def _load_json(path: Path) -> dict:
        with open(path) as f:
            return json.load(f)

    def get_tasks(self, domain: str = None, task_ids: list = None) -> list:
        """Get tasks filtered by domain or specific IDs."""
        if task_ids:
            id_set = set(task_ids)
            return [t for t in self.all_tasks if t["id"] in id_set]
        if domain:
            return [t for t in self.all_tasks if t["domain"] == domain.upper()]
        return list(self.all_tasks)

    def randomize_tasks(self, tasks: list, seed: int = None) -> list:
        """Randomize task order with optional seed for reproducibility."""
        if seed is None:
            seed = int(time.time())
        rng = random.Random(seed)
        shuffled = list(tasks)
        rng.shuffle(shuffled)
        return shuffled, seed

    def create_result_entry(self, task: dict, seed: int) -> dict:
        """Create a blank result entry for a task."""
        return {
            "experiment_id": self.config["experiment_id"],
            "group": self.group,
            "group_config": self.group_config["name"],
            "task_id": task["id"],
            "domain": task["domain"],
            "difficulty": task["difficulty"],
            "title": task["title"],
            "started_at": datetime.now(timezone.utc).isoformat(),
            "completed_at": None,
            "randomization_seed": seed,
            # Metrics (filled after execution)
            "completion_rate": None,
            "first_pass_rate": None,
            "avg_fix_rounds": 0,
            "error_density": None,
            "cross_domain_transfer": None,
            "time_seconds": None,
            "token_consumption": 0,
            # Execution metadata
            "retry_count": 0,
            "reviewer_notes": [],
            "raw_output_summary": "",
            "evaluation_notes": "",
            "status": "pending"  # pending -> executing -> completed/failed
        }

    def save_result(self, entry: dict):
        """Append a result entry to the results file."""
        with open(self.results_file, "a") as f:
            f.write(json.dumps(entry, ensure_ascii=False) + "\n")

    def load_existing_results(self) -> dict:
        """Load existing results to avoid re-executing completed tasks."""
        completed = {}
        if self.results_file.exists():
            with open(self.results_file) as f:
                for line in f:
                    line = line.strip()
                    if line:
                        entry = json.loads(line)
                        if entry["status"] == "completed":
                            completed[entry["task_id"]] = entry
        return completed

    def compute_summary(self):
        """Compute summary statistics from results file."""
        if not self.results_file.exists():
            return {"error": "No results file found"}

        entries = []
        with open(self.results_file) as f:
            for line in f:
                line = line.strip()
                if line:
                    entries.append(json.loads(line))

        completed = [e for e in entries if e["status"] == "completed"]
        if not completed:
            return {"error": "No completed results", "total_entries": len(entries)}

        summary = {
            "group": self.group,
            "group_name": self.group_config["name"],
            "architecture": self.group_config["architecture"],
            "total_tasks": len(entries),
            "completed": len(completed),
            "failed": len([e for e in entries if e["status"] == "failed"]),
            "pending": len([e for e in entries if e["status"] == "pending"]),
            # Aggregate metrics
            "completion_rate": sum(1 for e in completed if e["completion_rate"]) / len(completed) if completed else 0,
            "first_pass_rate": sum(1 for e in completed if e["first_pass_rate"]) / len(completed) if completed else 0,
            "avg_fix_rounds": sum(e["avg_fix_rounds"] for e in completed) / len(completed) if completed else 0,
            "avg_error_density": sum(e["error_density"] for e in completed if e["error_density"] is not None) / len([e for e in completed if e["error_density"] is not None]) if completed else 0,
            "avg_time_seconds": sum(e["time_seconds"] for e in completed if e["time_seconds"] is not None) / len([e for e in completed if e["time_seconds"] is not None]) if completed else 0,
            "total_token_consumption": sum(e["token_consumption"] for e in completed if e["token_consumption"]),
            # Per-domain breakdown
            "by_domain": {}
        }

        # Domain breakdown
        for entry in completed:
            d = entry["domain"]
            if d not in summary["by_domain"]:
                summary["by_domain"][d] = {"count": 0, "completed": 0, "completion_rate": 0}
            summary["by_domain"][d]["count"] += 1
            if entry["completion_rate"]:
                summary["by_domain"][d]["completed"] += 1

        for d in summary["by_domain"]:
            c = summary["by_domain"][d]
            c["completion_rate"] = c["completed"] / c["count"] if c["count"] > 0 else 0

        # Cross-domain transfer: compare first 5 vs last 5 for each domain
        by_domain_tasks = defaultdict(list)
        for e in completed:
            by_domain_tasks[e["domain"]].append(e)

        transfer_effects = {}
        for domain, domain_entries in by_domain_tasks.items():
            domain_entries.sort(key=lambda x: x["started_at"])
            if len(domain_entries) >= 10:
                first5 = domain_entries[:5]
                last5 = domain_entries[-5:]
                first5_rate = sum(1 for e in first5 if e["completion_rate"]) / 5
                last5_rate = sum(1 for e in last5 if e["completion_rate"]) / 5
                transfer_effects[domain] = last5_rate - first5_rate

        summary["cross_domain_transfer"] = transfer_effects

        # Save summary
        with open(self.summary_file, "w") as f:
            json.dump(summary, f, ensure_ascii=False, indent=2)

        return summary

    def prepare_batch(self, tasks: list, seed: int = None) -> tuple:
        """Prepare a batch of tasks for execution. Returns (new_tasks, seed)."""
        shuffled, seed = self.randomize_tasks(tasks, seed)

        # Create result entries for tracking
        existing = self.load_existing_results()
        new_entries = []

        for task in shuffled:
            if task["id"] not in existing:
                entry = self.create_result_entry(task, seed)
                entry["status"] = "pending"
                self.save_result(entry)
                new_entries.append(entry)

        return new_entries, seed

    def mark_started(self, task_id: str):
        """Mark a task as executing."""
        self._update_status(task_id, "executing")

    def mark_completed(self, task_id: str, metrics: dict):
        """Mark a task as completed with metrics."""
        self._update_status(task_id, "completed", metrics)

    def mark_failed(self, task_id: str, error: str):
        """Mark a task as failed."""
        self._update_status(task_id, "failed", {"error": error})

    def _update_status(self, task_id: str, status: str, extra: dict = None):
        """Update a task's status in the results file."""
        entries = []
        found = False
        if self.results_file.exists():
            with open(self.results_file) as f:
                for line in f:
                    line = line.strip()
                    if line:
                        entry = json.loads(line)
                        if entry["task_id"] == task_id:
                            entry["status"] = status
                            if status == "completed":
                                entry["completed_at"] = datetime.now(timezone.utc).isoformat()
                            if extra:
                                entry.update(extra)
                            found = True
                        entries.append(entry)

        if not found:
            raise ValueError(f"Task {task_id} not found in results")

        # Rewrite file
        with open(self.results_file, "w") as f:
            for entry in entries:
                f.write(json.dumps(entry, ensure_ascii=False) + "\n")


def print_experiment_plan(group: str, domain: str = None):
    """Print the execution plan without running."""
    runner = ExperimentRunner(group)
    tasks = runner.get_tasks(domain=domain)
    shuffled, seed = runner.randomize_tasks(tasks)
    print(f"Group {group} ({runner.group_config['name']})")
    print(f"Architecture: {runner.group_config['architecture']}")
    print(f"Components: {', '.join(runner.group_config['components'])}")
    print(f"Seed: {seed}")
    print(f"\nExecution plan ({len(shuffled)} tasks):")
    for i, t in enumerate(shuffled, 1):
        print(f"  {i:3d}. [{t['domain']}] {t['id']}: {t['title']} (difficulty={t['difficulty']})")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="AGI Cognitive Benchmark Experiment Runner")
    parser.add_argument("--group", required=True, choices=["A", "B", "C"], help="Experiment group")
    parser.add_argument("--domain", help="Filter by cognitive domain")
    parser.add_argument("--task-ids", help="Comma-separated task IDs")
    parser.add_argument("--prepare", action="store_true", help="Prepare batch without executing")
    parser.add_argument("--summary", action="store_true", help="Compute and print summary")
    parser.add_argument("--seed", type=int, help="Random seed for reproducibility")

    args = parser.parse_args()

    if args.summary:
        runner = ExperimentRunner(args.group)
        summary = runner.compute_summary()
        print(json.dumps(summary, ensure_ascii=False, indent=2))
        sys.exit(0)

    task_ids = args.task_ids.split(",") if args.task_ids else None
    print_experiment_plan(args.group, args.domain)

    if args.prepare:
        runner = ExperimentRunner(args.group)
        tasks = runner.get_tasks(domain=args.domain, task_ids=task_ids)
        entries, seed = runner.prepare_batch(tasks, args.seed)
        print(f"\nPrepared {len(entries)} new tasks (seed={seed})")
        print(f"Results file: {runner.results_file}")
