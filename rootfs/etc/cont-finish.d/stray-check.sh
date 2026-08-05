#!/bin/sh
# ==============================================================================
# Detect processes that escaped s6 supervision
# All s6-rc services have already been stopped gracefully at this point, so
# any process outside the s6 supervision tree was leaked by a service or
# init script. It is about to be SIGKILLed without any grace period
# (S6_KILL_GRACETIME=0): fix it by supervising the process, not by
# re-adding grace time.
# Plain sh with only builtins: bashio startup alone costs ~0.8s, which is
# most of the time this image saves by skipping the grace sleep.
# ==============================================================================

# A process is part of shutdown machinery if its ancestry (self included)
# reaches this script or an s6 process before reaching PID 1.
is_machinery() {
    local pid comm stat
    pid=$1
    while [ "${pid}" -gt 1 ] 2>/dev/null; do
        [ "${pid}" = "$$" ] && return 0
        IFS= read -r comm < "/proc/${pid}/comm" 2>/dev/null || return 0
        case "${comm}" in s6-*) return 0 ;; esac
        IFS= read -r stat < "/proc/${pid}/stat" 2>/dev/null || return 0
        set -- ${stat##*) }
        pid=$2
    done
    return 1
}

for proc in /proc/[0-9]*; do
    pid=${proc#/proc/}
    [ "${pid}" -eq 1 ] && continue

    IFS= read -r stat < "${proc}/stat" 2>/dev/null || continue
    # fields after the comm: $1 state, $2 ppid; skip zombies awaiting reaping
    set -- ${stat##*) }
    state=$1
    ppid=$2
    [ "${state}" = "Z" ] && continue

    is_machinery "${pid}" && continue

    IFS= read -r comm < "${proc}/comm" 2>/dev/null || continue
    echo "stray-check: WARNING: stray process at shutdown:" \
        "PID ${pid} (${comm}, parent ${ppid}) escaped s6 supervision" \
        "and is killed without grace period" >&2
done

exit 0
