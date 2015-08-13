# Actiontec V2200H → New Relic Insights

**Problem**: My bonded VDSL Internet service seems a bit crap.

**Added problem**: The Actiontec V2200H doesn't implement SNMP, because it is a
much bigger pile of crap.

**Solution**: A few hundred lines of crappy Go to screen scrape the admin
interface for the modem and push some basic line stats to New Relic Insights.

## Does it have go to Insights?

Nope. You could change `main.go` to be modular, or just hack it to push the
data wherever you want.

## Not all the stats I want are sent!

If they're on the modem status screen in the router UI, then they should be
accessible. You'll have to figure out what field they are and adjust the
`actiontec` module accordingly.

## I have a different Actiontec DSL modem. Will this work?

It might. I had a V1000H before this modem, and I suspect it was close enough
that you'd get something useful. Actiontec appear to reuse the same basic code
for their UI (which makes sense).

## What are some useful NRQL queries that I can put on an Insights dashboard?

These all assume you're interested in the last hour. Adjust accordingly if
you're not.

### Overall speed

    SELECT average(RateDown), average(RateUp)  FROM ModemStats TIMESERIES 1 minute SINCE 1 hour ago

### Line 1 speed

    SELECT average(RateDown), average(RateUp) FROM LineStats WHERE Line=0 TIMESERIES 1 minute SINCE 1 hour ago

### Line 1 attenuation

    SELECT average(AttenuationDown), average(AttenuationUp) FROM LineStats WHERE Line=0 TIMESERIES 1 minute SINCE 1 hour ago

### Line 1 SNR margin

    SELECT average(SignalNoiseMarginDown), average(SignalNoiseMarginUp) FROM LineStats WHERE Line=0 TIMESERIES 1 minute SINCE 1 hour ago

### Line 1 retrains

    SELECT max(Retrains) FROM LineStats WHERE Line=0 SINCE 1 hour ago

For line 2, just change `Line=0` to `Line=1` in the `LineStats` queries.
