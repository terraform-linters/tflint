require 'thor'
require 'aws-sdk'

class Wizard < Thor
  include Thor::Actions

  desc 'wizard generate', 'generate template for quick start'
  def generate
    region = 'us-east-1'
    client = Aws::EC2::Client.new(region: region)
    ec2 = Aws::EC2::Resource.new(client: client)

    default_vpc_id = ec2.vpcs(filters: [{ name: 'isDefault', values: [true.to_s] }]).first.id
    subnets = ec2.subnets(filters: [{ name: 'vpc-id', values: [default_vpc_id] }]).limit(2)
    key = ec2.key_pairs.first
    if key.nil?
      key = ec2.create_key_pair(key_name: 'demo-app')
      create_file 'demo-app.pem', key.key_material
    end
    spot_infos = []
    subnets.each do |subnet|
      res = client.describe_spot_price_history(instance_types: ['m3.medium'],
                                          product_descriptions: ["Linux/UNIX (Amazon VPC)"],
                                          availability_zone: subnet.availability_zone)
      spot_infos << { subnet: subnet.id, price: res.spot_price_history.first.spot_price }
    end

    template 'template.tf.erb', 'template.tf', { region: region, vpc_id: default_vpc_id, spot_infos: spot_infos, key_name: key.name }
  end

  desc 'wizard g', 'alias for wizard generate'
  alias_method :g, :generate
end

Wizard.source_root('.')
Wizard.start(ARGV)
