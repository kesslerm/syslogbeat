from syslogbeat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting Syslogbeat normally
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/*"
        )

        syslogbeat_proc = self.start_beat()
        self.wait_until(lambda: self.log_contains("syslogbeat is running"))
        exit_code = syslogbeat_proc.kill_and_wait()
        assert exit_code == 0
