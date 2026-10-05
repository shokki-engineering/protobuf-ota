# protobuf-ota

Protobuf types for the [OpenTravel Alliance](https://opentravel.org) hotel
specification, reshaped from OTA's XML request/response messages into
resource-oriented APIs that follow [Google's AIPs](https://aip.dev) and lint
clean under [api-linter](https://github.com/googleapis/api-linter).

It is built the way
[coversa-protobuf](https://github.com/oh-tarnished/coversa-protobuf) is: every
`.proto` is generated from a pinned revision of the upstream specification,
never written by hand, and no lint rule is disabled.

> **Status.** `just sync` writes 28 resources -- standard, custom and batch
> methods, and a README per package -- into `protobuf/`: 1504 protos in 29
> packages, with zero api-linter findings (2026-10-05). Start with [`CLAUDE.md`](CLAUDE.md) for
> the rules and [`docs/resources.md`](docs/resources.md) for what is
> generated; [`protobuf/opentravel/README.md`](protobuf/opentravel/README.md)
> is the generated index.

## What it models

The OTA 2017B XML message suite -- hotel availability and booking, rates and
inventory, descriptive content, group blocks, events, RFPs and statistics --
with codes drawn from the OpenTravel code list of 2024-09-04. Both come from
[OpenTravel-Specifications](https://github.com/OpenTravel/OpenTravel-Specifications),
checked out at `modules/opentravel`.

An OTA message becomes an AIP method on a resource:

| OTA | Here |
| --- | --- |
| `OTA_HotelResRQ` | `Reservations.CreateReservation` |
| `OTA_HotelResModifyRQ` | `Reservations.UpdateReservation` |
| `OTA_CancelRQ` | `Reservations.CancelReservation` |
| `OTA_HotelRatePlanNotifRQ` | `RatePlans.CreateRatePlan` / `UpdateRatePlan` |
| `OTA_HotelAvailNotifRQ` | `Restrictions.BatchUpdateRestrictions` |
| `OTA_HotelAvailRQ` | `Hotels.SearchAvailability` |

## Layout

```text
protobuf/opentravel/<spec>/<domain>/<resource>/
    README.md                its generated reference
    v1/                      package ...<resource>.v1
        <resource>.proto     the resource
        messages.proto       its standard requests and responses
        service.proto        its methods
        <custom_method>.proto  a custom method's request and response
        types/               ...v1.types: OTA value messages, Money, Date
        codes/               ...v1.codes: OpenTravel code-table enums
        enums/               ...v1.enums: closed XSD enumerations
```

`<spec>` is `hotel` or `common`; `<domain>` is one of `property`, `revenue`,
`booking`, `sales`, `finance`, `reporting`. One top-level message or enum per
file, except a cycle of types, which protobuf makes share one.

## Documentation

| | |
| --- | --- |
| [`CLAUDE.md`](CLAUDE.md) | the working rules, 1-9 |
| [`docs/conventions.md`](docs/conventions.md) | rules 10-15 and the rename catalogue |
| [`docs/resources.md`](docs/resources.md) | every resource, and the method each OTA message becomes |
| [`docs/decisions.md`](docs/decisions.md) | what was deliberately not done, and why |
| [`docs/decisions-generator.md`](docs/decisions-generator.md) | how the XSD's shapes become protobuf |
| [`docs/spec.md`](docs/spec.md) | the pinned revision and how to move it |
| [`docs/generator.md`](docs/generator.md) | how `sync` and `docs` are laid out |
| [`docs/references.md`](docs/references.md) | a checked link index |

## Cloning

```sh
git clone --recurse-submodules <this repository>
```

The submodule is large -- OpenTravel's archive holds every publication since
2001 -- and only two directories of it are read.
