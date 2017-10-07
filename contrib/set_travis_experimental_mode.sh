# Enable pre-release docker features in the travis build.
# Can't be in .travis.yml as travis has trouble parsing the braces.
echo "{\"experimental\": true}" | sudo tee /etc/docker/daemon.json
