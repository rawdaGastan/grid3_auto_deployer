rm -rf /root/c4s-backup/*

# export restic variables
export AWS_ACCESS_KEY_ID="<access_key_id>"
export AWS_SECRET_ACCESS_KEY="<secret_access_key>"
export RESTIC_PASSWORD="<password>"
export RESTIC_REPOSITORY="<repository>"

# take database dump to the host
cp database.sql /root/c4s-backup/c4s-db-backup-$(date +"%m-%d-%Y"+"%T").sql

# upload the backup directory to s3 backup server
/usr/bin/restic backup /root/c4s-backup/* 2> /var/log/restic/restic.err > /var/log/restic/restic.log
