// Code generated from formulas.Schemas(); DO NOT EDIT.

export const FORMULA_SCHEMA_NAME = "ProductionFormulasV14" as const;
export const FORMULA_SCHEMAS = [
  {
    "name": "ProductionFormulaAxisStack",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "factor",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "input",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "pinned",
          "required": true,
          "schema": {
            "kind": "oneOf",
            "alternates": [
              {
                "kind": "null"
              },
              {
                "kind": "ref",
                "ref": "ProductionFormulaAxisStackPinning"
              }
            ]
          }
        },
        {
          "name": "slot",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "stack",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "timing",
          "required": true,
          "schema": {
            "kind": "string"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaAxisStackPinning",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "cap_reason_key",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "mechanical-id"
          }
        },
        {
          "name": "input",
          "required": true,
          "schema": {
            "kind": "string",
            "enum": [
              "achievement_attainment_run",
              "achievement_score_run",
              "clout_run"
            ]
          }
        },
        {
          "name": "input_cap",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "maximum_product_at_cap",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "canonical-decimal"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaCommons",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "cohort_health_weight_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "cohort_merge_floor",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "cohort_target_size",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "collapse_health_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "collective_exponent_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "collective_weight_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "compliance",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "default_tithe_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "effective_health",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "enclosure",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "entry_participation_weight",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "guild_health_weight_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "health",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "health_decay_ppm_per_hour",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "health_recovery_ppm_per_hour",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "healthy_health_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "maximum_bonus",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "canonical-decimal"
          }
        },
        {
          "name": "maximum_tithe_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "minimum_tithe_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "modifier",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "npc_compliance_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "npc_population_floor",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "npc_weight_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "population_tolerance_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "server_health_weight_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "solidarity",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "solidarity_window_ms",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "source_weights",
          "required": true,
          "schema": {
            "kind": "array",
            "items": {
              "kind": "ref",
              "ref": "ProductionFormulaSourceWeight"
            }
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaGuild",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "clearing",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "clearing_interval_ms",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "clearing_rate_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "consumption_bonus_ppm_per_unit",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "guild_tithe_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "guild_xp_target_per_founder",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "health",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "npc_exchange_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "stock_consumption",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "stock_intake_cap",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "tithe",
          "required": true,
          "schema": {
            "kind": "string"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaHardcap",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "count",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "generator_id",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "mechanical-id"
          }
        },
        {
          "name": "reason_key",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "mechanical-id"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaMeters",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "attended_step",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "band_events",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "contribution_input",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "decay",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "hook_order",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "ledger_fact_input",
          "required": true,
          "schema": {
            "kind": "string"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaMinigameScaling",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "fairness_gate",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "fallback_arms",
          "required": true,
          "schema": {
            "kind": "array",
            "items": {
              "kind": "string",
              "enum": [
                "bot",
                "npc_partner",
                "solo"
              ]
            }
          }
        },
        {
          "name": "fallback_rule",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "grammar",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "offline_grade_grammar",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "offline_grade_rule",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "operation_order",
          "required": true,
          "schema": {
            "kind": "array",
            "items": {
              "kind": "string"
            }
          }
        },
        {
          "name": "payout_grammar",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "payout_math",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "payout_order",
          "required": true,
          "schema": {
            "kind": "array",
            "items": {
              "kind": "string"
            }
          }
        },
        {
          "name": "rounding",
          "required": true,
          "schema": {
            "kind": "string"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaPurchasableContent",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "ladders",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "manual_output",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "provision_tick_ms",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        },
        {
          "name": "provisioned_hardcaps",
          "required": true,
          "schema": {
            "kind": "array",
            "items": {
              "kind": "ref",
              "ref": "ProductionFormulaHardcap"
            }
          }
        },
        {
          "name": "provisioning",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "stock_rate",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "synergy_linear",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "synergy_log",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "synergy_pools",
          "required": true,
          "schema": {
            "kind": "array",
            "items": {
              "kind": "ref",
              "ref": "ProductionFormulaSynergyPool"
            }
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaSourceWeight",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "forsworn",
          "required": true,
          "schema": {
            "kind": "boolean"
          }
        },
        {
          "name": "slot",
          "required": true,
          "schema": {
            "kind": "string",
            "enum": [
              "axis_stack",
              "commons",
              "doctrine",
              "event_buffs",
              "faction",
              "milestones",
              "prestige",
              "trust",
              "upgrades"
            ]
          }
        },
        {
          "name": "source_id",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "mechanical-id"
          }
        },
        {
          "name": "weight_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaSynergyPool",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "curve",
          "required": true,
          "schema": {
            "kind": "string",
            "enum": [
              "linear",
              "log"
            ]
          }
        },
        {
          "name": "id",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "mechanical-id"
          }
        },
        {
          "name": "slot",
          "required": true,
          "schema": {
            "kind": "string",
            "enum": [
              "axis_stack",
              "commons",
              "doctrine",
              "event_buffs",
              "faction",
              "milestones",
              "prestige",
              "trust",
              "upgrades"
            ]
          }
        },
        {
          "name": "sources",
          "required": true,
          "schema": {
            "kind": "array",
            "items": {
              "kind": "ref",
              "ref": "ProductionFormulaSynergySource"
            }
          }
        },
        {
          "name": "target",
          "required": true,
          "schema": {
            "kind": "string"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulaSynergySource",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "id_or_class",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "mechanical-id"
          }
        },
        {
          "name": "kind",
          "required": true,
          "schema": {
            "kind": "string",
            "enum": [
              "generator",
              "upgrade"
            ]
          }
        },
        {
          "name": "per_count_ppm",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "-9223372036854775808",
            "maximum": "9223372036854775807"
          }
        }
      ]
    }
  },
  {
    "name": "ProductionFormulasV14",
    "schema": {
      "kind": "object",
      "fields": [
        {
          "name": "axis_stack",
          "required": true,
          "schema": {
            "kind": "ref",
            "ref": "ProductionFormulaAxisStack"
          }
        },
        {
          "name": "commons",
          "required": true,
          "schema": {
            "kind": "ref",
            "ref": "ProductionFormulaCommons"
          }
        },
        {
          "name": "guild",
          "required": true,
          "schema": {
            "kind": "ref",
            "ref": "ProductionFormulaGuild"
          }
        },
        {
          "name": "meters",
          "required": true,
          "schema": {
            "kind": "ref",
            "ref": "ProductionFormulaMeters"
          }
        },
        {
          "name": "minigame_scaling",
          "required": true,
          "schema": {
            "kind": "ref",
            "ref": "ProductionFormulaMinigameScaling"
          }
        },
        {
          "name": "multiplier_slot_order",
          "required": true,
          "schema": {
            "kind": "array",
            "items": {
              "kind": "string",
              "enum": [
                "axis_stack",
                "commons",
                "doctrine",
                "event_buffs",
                "faction",
                "milestones",
                "prestige",
                "trust",
                "upgrades"
              ]
            }
          }
        },
        {
          "name": "production_rate",
          "required": true,
          "schema": {
            "kind": "string"
          }
        },
        {
          "name": "purchasable_content",
          "required": true,
          "schema": {
            "kind": "ref",
            "ref": "ProductionFormulaPurchasableContent"
          }
        },
        {
          "name": "schema_version",
          "required": true,
          "schema": {
            "kind": "integer",
            "minimum": "14",
            "maximum": "14"
          }
        },
        {
          "name": "source_fingerprint",
          "required": true,
          "schema": {
            "kind": "string",
            "format": "sha256"
          }
        },
        {
          "name": "within_slot_order",
          "required": true,
          "schema": {
            "kind": "string",
            "enum": [
              "source_id_raw_byte_ascending"
            ]
          }
        }
      ]
    }
  }
] as const;
