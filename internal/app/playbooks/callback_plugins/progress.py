import time
from ansible.plugins.callback import CallbackBase


# ANSI 색상 코드 정의
GREEN = '\033[92m'
RED = '\033[91m'
RESET = '\033[0m'
WHITE = '\033[97m'


class CallbackModule(CallbackBase):
    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'aggregate'
    CALLBACK_NAME = 'progress'

    def __init__(self):
        super().__init__()
        self.current_play_index = 0
        self.total_plays = 0
        self.start_time = time.time()
        self.current_play = None
        self.play_start_time = None
        self.play_tasks = {}
        self.total_tasks = 0
        self.executed_tasks = 0
        self.executed_plays = 0
        self.all_roles = []
        self.role_execution_times = {}
        self.executed_roles = {}
        self.skipped_roles = set()
        self.errors = []
        self.warnings = []

    def format_time(self, seconds):
        hours, remainder = divmod(int(seconds), 3600)
        minutes, seconds = divmod(remainder, 60)
        if hours > 0:
            return f"{hours}h {minutes}m {seconds}s"
        elif minutes > 0:
            return f"{minutes}m {seconds}s"
        else:
            return f"{seconds}s"
    def v2_playbook_on_start(self, playbook):
        self.total_plays = len(playbook.get_plays())

        executed_roles = []
        skipped_roles = []

        for idx, play in enumerate(playbook.get_plays(), start=1):
            role_name = play.get_name()
            if role_name not in self.all_roles:
                self.all_roles.append(role_name)
                if self.will_execute_role(play):
                    executed_roles.append(f"{idx}. {role_name}")
                else:
                    skipped_roles.append(f"{idx}. {role_name}")

    def will_execute_role(self, play):
        when = play.get_vars().get('when')
        if when is None:
            return True
        return str(when).lower() == 'true'

    def v2_playbook_on_play_start(self, play):
        self.current_play_index += 1
        self.current_play = play.get_name()
        self.play_start_time = time.time()
        self.play_tasks[self.current_play] = 0
        if self.current_play not in self.executed_roles:
            self.executed_roles[self.current_play] = 0

    def v2_playbook_on_task_start(self, task, is_conditional):
        self.play_tasks[self.current_play] += 1
        self.total_tasks += 1
        if self.executed_plays < self.current_play_index:
            self.executed_plays = self.current_play_index
        elapsed_time = time.time() - self.start_time
        print(f"Play ({self.current_play_index}/{self.total_plays}), "
              f"Role: {self.current_play}, "
              f"Task {self.play_tasks[self.current_play]}, "
              f"Total Time: {self.format_time(elapsed_time)}")

    def v2_runner_on_ok(self, result):
        self.executed_tasks += 1
        self.executed_roles[self.current_play] += 1
        self._update_role_execution_time()

    def v2_runner_on_skipped(self, result):
        self.skipped_roles.add(self.current_play)
        self._update_role_execution_time()

    def v2_runner_on_failed(self, result, ignore_errors=False):
        self.executed_tasks += 1
        if not ignore_errors:
            error_msg = f"Error in role {self.current_play}: {result._result.get('msg', 'unknown error')}"
            self.errors.append(error_msg)
            print(f"\n{error_msg}")
        self._update_role_execution_time()

    def v2_runner_on_unreachable(self, result):
        self.executed_tasks += 1
        error_msg = f"Host unreachable in role {self.current_play}: {result._result.get('msg', 'unknown error')}"
        self.errors.append(error_msg)
        print(f"\n{error_msg}")
        self._update_role_execution_time()

    def v2_runner_on_warning(self, host_name, warning):
        self.warnings.append(f"Warning in {self.current_play}: {warning}")
        self._update_role_execution_time()

    def v2_playbook_on_play_end(self, play):
        self._update_role_execution_time()

    def _update_role_execution_time(self):
        if self.current_play:
            end_time = time.time()
            duration = end_time - self.play_start_time
            self.role_execution_times[self.current_play] = duration

    def v2_playbook_on_stats(self, stats):
        total_duration = time.time() - self.start_time

        print("\nPlaybook Execution Summary:")

        print(f"Total Duration: {self.format_time(total_duration)}")
        print(f"Total Executed Tasks: {self.executed_tasks}")
        print(f"Executed Roles: ")
        for role, count in self.executed_roles.items():
            duration = self.role_execution_times.get(role, 0)
            if count > 0:
                color = GREEN
            else:
                color = WHITE
            print(f"{color}- {role}: {count} tasks, {self.format_time(duration)}{RESET}")

        if self.skipped_roles:
            print("")
            for role in self.skipped_roles:
                print(f"")

        if self.warnings:
            print("\nWarnings:")
            for warning in self.warnings:
                print(f"- {warning}")

        if self.errors:
            print("\nErrors:")
            for error in self.errors:
                print(f"{RED}- {error}{RESET}")

        success_rate = (self.executed_tasks - len(self.errors)) / self.executed_tasks * 100 if self.executed_tasks > 0 else 0
        print(f"\nSuccess Rate: {success_rate:.2f}%")
