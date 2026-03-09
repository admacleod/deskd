use warnings;
use strict;

use FindBin '$Bin';
use Cwd 'abs_path';
use Test::More;
use File::Temp;

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
    'expect_booking',
    'expect_response'
);

sub read_response {
    my ($result) = @_;
    my %response = (
        'headers' => [],
        'html'    => '',
    );
    my $reading_headers = 1;
    my @html = ();
    while ( my $line = <$result> ) {
        # Remove \r\n and \n regardless of the value of $/ (used by chomp)
        # Some CGI implementations seem to use \r\n for Headers and then \n for HTML.
        $line =~ s/\r?\n//;
        if ( $reading_headers != 0 ) {
            if ( $line ne '' ) {
                push(@{$response{'headers'}}, $line);
            } else {
                $reading_headers = 0;
            }
        } else {
            push(@html, $line);
        }
    }
    $response{'html'} = join("\n", @html);

    # Remove any trailing whitespace from the HTML which can prevent matching in tests and isn't exactly relevant.
    $response{'html'} =~ s/\n+\z//;

    return(\%response);
}

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
        # Read the expected response.
        open(my $expect_response, '<', \$test{expect_response}) or die 'failed to open expected response for reading.';
        my %expect = %{read_response($expect_response)};

        # Go to a new temporary directory.
        my $temp_dir = File::Temp->newdir();
        my $db_path = "$temp_dir/deskd.db";

        # Initialise the database in the new directory.
        local %ENV = (%ENV,
            DESKD_DB => $db_path
        );
        system($deskd_bin,  'migrate') == 0 or die 'failed to run database migration.';

        # Construct the environment to run in.
        my %child_env = (%ENV);
        foreach my $line ( split("\n", $test{request_env}) ) {
            next if $line eq '';
            my ($key, $value) = split(/=/, $line, 2);
            $child_env{$key} = $value;
        }
        local %ENV = %child_env;

        my $pid = open(my $result, '-|');
        die "failed to fork: $!" unless defined $pid;
        if ( $pid == 0 ) {
            # Redirect STDERR as we don't want it to pollute TAP output on purposefully erroring test cases.
            open(STDERR, '>', '/dev/null') or die "failed to redirect STDERR: $!";
            exec $deskd_bin or die "failed to exec $deskd_bin: $!";
        }

        my %response = %{read_response($result)};
        close($result);
        is($? >> 8, 0, 'deskd exited successfully');

        is_deeply($response{'headers'}, $expect{'headers'}, 'headers match');
        is($response{'html'}, $expect{'html'}, 'html matches');
    };
}
