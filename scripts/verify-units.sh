#!/usr/bin/env bash
# =============================================================================
# verify-units.sh — проверка systemd-юнитов и desktop-файла (E7/7.6)
# =============================================================================
# Статическая валидация:
#   - наличие секций и обязательных ключей systemd-unit;
#   - согласованность .timer и .service (Unit= указывает на существующий файл);
#   - корректность desktop-файла (Type/Name/Exec/TryExec/Icon);
#   - отсутствие заведомо нерабочих настроек (ExecReload без обработчика
#     сигналов, MimeType-схемы, неверный URL документации).
#
# Запуск: ./scripts/verify-units.sh  (или make units-check)
# =============================================================================

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
SYSTEMD_DIR="$ROOT_DIR/config/systemd"
DESKTOP_FILE="$ROOT_DIR/config/desktop/network-scanner-gui.desktop"

ERRORS=0
CHECKS=0

fail() {
    ERRORS=$((ERRORS + 1))
    echo "  [FAIL] $1"
}

ok() {
    CHECKS=$((CHECKS + 1))
    echo "  [ok]   $1"
}

section_exists() {
    grep -qE "^\[$2\]" "$1"
}

key_exists() {
    grep -qE "^$2=" "$1"
}

key_value() {
    grep -E "^$2=" "$1" | head -1 | cut -d= -f2-
}

echo "=== systemd unit validation ==="

SERVICE_FILE="$SYSTEMD_DIR/network-scanner-scan.service"
TIMER_FILE="$SYSTEMD_DIR/network-scanner-scan.timer"

# --- .service ---
if [ ! -f "$SERVICE_FILE" ]; then
    fail "missing $SERVICE_FILE"
else
    ok "found $(basename "$SERVICE_FILE")"

    for section in Unit Service Install; do
        if section_exists "$SERVICE_FILE" "$section"; then
            ok "section [$section]"
        else
            fail "missing section [$section] in $(basename "$SERVICE_FILE")"
        fi
    done

    for key in Description Documentation After ExecStart Type User Group; do
        if key_exists "$SERVICE_FILE" "$key"; then
            ok "key $key"
        else
            fail "missing key $key in $(basename "$SERVICE_FILE")"
        fi
    done

    # ExecStart должен указывать на установленный бинарь и реальную подкоманду.
    exec_start="$(key_value "$SERVICE_FILE" ExecStart)"
    case "$exec_start" in
        /usr/local/bin/network-scanner\ scan*)
            ok "ExecStart invokes installed binary with subcommand"
            ;;
        *)
            fail "ExecStart must call /usr/local/bin/network-scanner scan: '$exec_start'"
            ;;
    esac

    # oneshot: сканирование разовое, долгоживущий процесс здесь неуместен.
    if [ "$(key_value "$SERVICE_FILE" Type)" = "oneshot" ]; then
        ok "Type=oneshot (разовое сканирование)"
    else
        fail "Type should be oneshot for a scan job, got '$(key_value "$SERVICE_FILE" Type)'"
    fi

    # ExecReload без обработчика сигналов — ложная возможность.
    if key_exists "$SERVICE_FILE" ExecReload; then
        fail "ExecReload present but app has no SIGHUP handler"
    else
        ok "no bogus ExecReload"
    fi

    if [ "$(key_value "$SERVICE_FILE" User)" = "root" ]; then
        fail "User=root: prefer dedicated user with capabilities"
    else
        ok "dedicated user ($(key_value "$SERVICE_FILE" User))"
    fi

    for key in NoNewPrivileges ProtectSystem ReadWritePaths; do
        if key_exists "$SERVICE_FILE" "$key"; then
            ok "$key"
        else
            fail "missing $key"
        fi
    done

    rw="$(key_value "$SERVICE_FILE" ReadWritePaths)"
    case "$rw" in
        *"/var/lib/network-scanner"*) ok "ReadWritePaths includes /var/lib/network-scanner" ;;
        *) fail "ReadWritePaths missing /var/lib/network-scanner: '$rw'" ;;
    esac

    doc="$(key_value "$SERVICE_FILE" Documentation)"
    case "$doc" in
        https://github.com/RekadzeAV/network-scanner*)
            ok "Documentation points to the real repository"
            ;;
        *)
            fail "Documentation URL must be https://github.com/RekadzeAV/network-scanner, got '$doc'"
            ;;
    esac
fi

# --- .timer ---
if [ ! -f "$TIMER_FILE" ]; then
    fail "missing $TIMER_FILE"
else
    ok "found $(basename "$TIMER_FILE")"
    for section in Timer Install; do
        if section_exists "$TIMER_FILE" "$section"; then
            ok "section [$section]"
        else
            fail "missing section [$section] in $(basename "$TIMER_FILE")"
        fi
    done
    for key in OnBootSec OnUnitActiveSec Unit WantedBy; do
        if key_exists "$TIMER_FILE" "$key"; then
            ok "timer key $key"
        else
            fail "missing timer key $key"
        fi
    done

    # Timer.Unit должен соответствовать существующему .service.
    unit_ref="$(key_value "$TIMER_FILE" Unit)"
    if [ -f "$SYSTEMD_DIR/$unit_ref" ]; then
        ok "timer references existing unit: $unit_ref"
    else
        fail "timer Unit='$unit_ref' does not match any file in $SYSTEMD_DIR"
    fi
fi

# --- desktop file ---
echo "=== desktop entry validation ==="
if [ ! -f "$DESKTOP_FILE" ]; then
    fail "missing $DESKTOP_FILE"
else
    ok "found $(basename "$DESKTOP_FILE")"

    if head -1 "$DESKTOP_FILE" | grep -q "^\[Desktop Entry\]"; then
        ok "[Desktop Entry] header present"
    else
        fail "first line must be [Desktop Entry]"
    fi

    for key in Type Name Exec Icon TryExec; do
        if key_exists "$DESKTOP_FILE" "$key"; then
            ok "desktop key $key"
        else
            fail "missing desktop key $key"
        fi
    done

    d_exec="$(key_value "$DESKTOP_FILE" Exec)"
    d_try="$(key_value "$DESKTOP_FILE" TryExec)"
    if [ "$d_exec" = "/usr/local/bin/network-scanner-gui" ]; then
        ok "Exec points to installed GUI binary"
    else
        fail "Exec should be /usr/local/bin/network-scanner-gui, got '$d_exec'"
    fi
    if [ "$d_try" = "$d_exec" ]; then
        ok "TryExec matches Exec"
    else
        fail "TryExec should match Exec"
    fi

    # Приложение не регистрирует схему network:// — объявление даёт
    # неработающую ассоциацию в браузерах и меню «Открыть с помощью».
    if key_value "$DESKTOP_FILE" MimeType" | grep -q "x-scheme-handler/network"; then
        fail "declares x-scheme-handler/network but app does not register the scheme"
    else
        ok "no unsupported scheme handler"
    fi

    icon_name="$(key_value "$DESKTOP_FILE" Icon)"
    if [ -f "$ROOT_DIR/assets/icons/$icon_name.svg" ]; then
        ok "icon asset exists: assets/icons/$icon_name.svg"
    else
        fail "icon asset assets/icons/$icon_name.svg not found"
    fi
fi

# --- итог ---
echo "=== result ==="
echo "checks passed: $CHECKS"
if [ "$ERRORS" -gt 0 ]; then
    echo "errors: $ERRORS"
    exit 1
fi
echo "errors: 0"
exit 0