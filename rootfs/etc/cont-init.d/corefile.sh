#!/usr/bin/with-contenv bashio
# ==============================================================================
# Handle corefile generation
# ==============================================================================

# MIGRATION: can removed later
if [ ! -f /config/coredns.json ]; then
    bashio::log.warning "Run old fashion"
    cp -f /config/corefile /etc/corefile

    bashio::exit.ok
elif [ -f /config/corefile ]; then
    bashio::log.info "Cleanup old corefile"
    rm -f /config/corefile
fi

# Generate corefile
if ! tempio \
        -conf /config/coredns.json \
        -template /usr/share/corefile.tempio \
        -out /etc/corefile
then
    bashio::log.error "Corefile fails to generate. Use fallback corefile"
fi
