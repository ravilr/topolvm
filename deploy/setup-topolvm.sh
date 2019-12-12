mkdir -p /tmp/topolvm/lvmd
truncate --size=20G /tmp/backing_store
sudo losetup -f /tmp/backing_store
sudo vgcreate -y myvg $(sudo losetup -j /tmp/backing_store | cut -d: -f1)
cat > lvmd.service <<EOF
[Unit]
Description=lvmd for TopoLVM
Wants=lvm2-monitor.service
After=lvm2-monitor.service

[Service]
Type=simple
Restart=on-failure
RestartForceExitStatus=SIGPIPE
ExecStartPre=/bin/mkdir -p /run/topolvm
ExecStart=/opt/sbin/lvmd --volume-group=myvg --listen=/run/topolvm/lvmd.sock

[Install]
WantedBy=multi-user.target
EOF
sudo systemd-run --unit=lvmd.service /home/core/lvmd --volume-group=myvg --listen=/run/topolvm/lvmd.sock --spare=1
