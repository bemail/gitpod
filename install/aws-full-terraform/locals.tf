locals {
  storage_vals = <<-EOF
  %{ if length(module.storage) != 0 }
    ${module.storage[0].values}
  %{ else }
  %{endif}
  EOF
  db_vals = <<-EOF
  %{ if length(module.database) != 0 }
    ${module.database[0].values}
  %{ else }
  %{endif}
  EOF
}