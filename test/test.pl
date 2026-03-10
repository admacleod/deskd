# Copyright 2026 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
#
# This file is part of deskd.
#
# deskd is free software: you can redistribute it and/or modify it under the
# terms of the GNU Affero General Public License as published by the Free
# Software Foundation, either version 3 of the License, or (at your option) any
# later version.
#
# deskd is distributed in the hope that it will be useful, but WITHOUT ANY
# WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
# A PARTICULAR PURPOSE. See the GNU Affero General Public License for more
# details.
#
# You should have received a copy of the GNU Affero General Public License
# along with deskd. If not, see <https://www.gnu.org/licenses/>.

use warnings;
use strict;

use FindBin '$Bin';
use Cwd 'abs_path';
use Symbol 'gensym';
use Test::More;
use File::Temp;
use IPC::Open3;

my ($deskd_bin) = @ARGV;
if ( !defined($deskd_bin) ) {
    die 'You must pass the location of the deskd binary to be tested!';
}
$deskd_bin = abs_path($deskd_bin);

my @test_format = (
    'name',
    'desk_seed',
    'booking_seed',
    'request_env',
    'request_data',
    'expect_booking',
    'expect_response'
);

my @tests = sort glob("$Bin/*.test");
plan tests => scalar @tests;
foreach ( @tests ) {
    open(my $fh, '<', $_) or die $!;
    my $i = 0;
    my %test = ();
    my @section = ();
    while ( my $line = <$fh> ) {
        chomp($line);
        if ( $line eq '---' ) {
            $test{$test_format[$i]} = join("\n", @section);
            @section = ();
            $i++;
            next;
        }
        push(@section, $line);
    }
    $test{$test_format[$i]} = join("\n", @section);

    subtest "$test{name} ($_)" => sub {
        # Go to a new temporary directory.
        my $temp_dir = File::Temp->newdir();
        my $db_path = "$temp_dir/deskd.db";

        # Initialise the database in the new directory.
        local %ENV = (%ENV,
            DESKD_DB => $db_path
        );
        system($deskd_bin,  'migrate') == 0 or die 'failed to run database migration.';

        # Seed the database.
        my @sql = ('PRAGMA foreign_keys = ON;');
        foreach my $line ( split("\n", $test{desk_seed}) ) {
            next if $line eq '';
            push(@sql, "INSERT INTO desks (desk) VALUES ('$line');");
        }
        foreach my $line ( split("\n", $test{booking_seed}) ) {
            next if $line eq '';
            my ($user, $desk, $day) = split(/,/, $line, 3);
            push(@sql, "INSERT INTO bookings (user, desk, day) VALUES ('$user', '$desk', '$day');");
        }
        my $sqlite_error = gensym();
        my $sqlite_pid = open3(my $sqlite_input, my $sqlite_output, $sqlite_error, 'sqlite3', $db_path);
        print {$sqlite_input} join("\n", @sql) . "\n";
        close($sqlite_input);
        waitpid($sqlite_pid, 0);
        if ( ($? >> 8) != 0 ) {
            die 'failed to execute sqlite command';
        }

        # Construct the environment to run in.
        my %child_env = (%ENV);
        foreach my $line ( split("\n", $test{request_env}) ) {
            next if $line eq '';
            my ($key, $value) = split(/=/, $line, 2);
            $child_env{$key} = $value;
        }
        local %ENV = %child_env;

        my $error = gensym();
        my $pid = open3(my $input, my $result, $error, $deskd_bin);
        print {$input} $test{request_data};
        close($input);
        my $response = do { local $/; <$result> };
        my $stderr = do { local $/; <$error> };
        close($result);
        close($error);
        waitpid($pid, 0);
        is($? >> 8, 0, 'deskd exited successfully');

        # Normalize line breaks so that it will match test case.
        $response =~ s/\r\n/\n/g;
        $response =~ s/\n+\z//;
        like($response, qr/\A$test{'expect_response'}\z/s, 'response matches');

        # Compare database state after run.
        my $sqlite_actual_error = gensym();
        my $sqlite_actual_pid = open3(my $sqlite_actual_input, my $sqlite_actual_output, $sqlite_actual_error, 'sqlite3', $db_path, '-batch', '-noheader', '-separator', ',');
        print {$sqlite_actual_input} 'SELECT user, desk, day FROM bookings ORDER BY user, desk, day;';
        close($sqlite_actual_input);
        my $sqlite_actual = do { local $/; <$sqlite_actual_output> };
        close($sqlite_actual_output);
        waitpid($sqlite_actual_pid, 0);
        if ( ($? >> 8) != 0 ) {
            die 'failed to execute sqlite command after run';
        }
        # Normalize line breaks so that it will match test case.
        $sqlite_actual =~ s/\r\n/\n/g;
        $sqlite_actual =~ s/\n+\z//;
        is($sqlite_actual, $test{expect_booking}, 'bookings match');

        diag($stderr) if !Test::More->builder->is_passing && $stderr ne '';
    };
}
