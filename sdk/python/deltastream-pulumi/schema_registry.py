# coding=utf-8
# *** WARNING: this file was generated by pulumi-language-python. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import builtins
import copy
import warnings
import sys
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
if sys.version_info >= (3, 11):
    from typing import NotRequired, TypedDict, TypeAlias
else:
    from typing_extensions import NotRequired, TypedDict, TypeAlias
from . import _utilities
from . import outputs
from ._inputs import *

__all__ = ['SchemaRegistryArgs', 'SchemaRegistry']

@pulumi.input_type
class SchemaRegistryArgs:
    def __init__(__self__, *,
                 confluent: Optional[pulumi.Input['SchemaRegistryConfluentArgs']] = None,
                 confluent_cloud: Optional[pulumi.Input['SchemaRegistryConfluentCloudArgs']] = None,
                 name: Optional[pulumi.Input[builtins.str]] = None,
                 owner: Optional[pulumi.Input[builtins.str]] = None):
        """
        The set of arguments for constructing a SchemaRegistry resource.
        :param pulumi.Input['SchemaRegistryConfluentArgs'] confluent: Confluent specific configuration
        :param pulumi.Input['SchemaRegistryConfluentCloudArgs'] confluent_cloud: Confluent cloud specific configuration
        :param pulumi.Input[builtins.str] name: Name of the schema registry
        :param pulumi.Input[builtins.str] owner: Owning role of the schema registry
        """
        if confluent is not None:
            pulumi.set(__self__, "confluent", confluent)
        if confluent_cloud is not None:
            pulumi.set(__self__, "confluent_cloud", confluent_cloud)
        if name is not None:
            pulumi.set(__self__, "name", name)
        if owner is not None:
            pulumi.set(__self__, "owner", owner)

    @property
    @pulumi.getter
    def confluent(self) -> Optional[pulumi.Input['SchemaRegistryConfluentArgs']]:
        """
        Confluent specific configuration
        """
        return pulumi.get(self, "confluent")

    @confluent.setter
    def confluent(self, value: Optional[pulumi.Input['SchemaRegistryConfluentArgs']]):
        pulumi.set(self, "confluent", value)

    @property
    @pulumi.getter(name="confluentCloud")
    def confluent_cloud(self) -> Optional[pulumi.Input['SchemaRegistryConfluentCloudArgs']]:
        """
        Confluent cloud specific configuration
        """
        return pulumi.get(self, "confluent_cloud")

    @confluent_cloud.setter
    def confluent_cloud(self, value: Optional[pulumi.Input['SchemaRegistryConfluentCloudArgs']]):
        pulumi.set(self, "confluent_cloud", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Name of the schema registry
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter
    def owner(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Owning role of the schema registry
        """
        return pulumi.get(self, "owner")

    @owner.setter
    def owner(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "owner", value)


@pulumi.input_type
class _SchemaRegistryState:
    def __init__(__self__, *,
                 confluent: Optional[pulumi.Input['SchemaRegistryConfluentArgs']] = None,
                 confluent_cloud: Optional[pulumi.Input['SchemaRegistryConfluentCloudArgs']] = None,
                 created_at: Optional[pulumi.Input[builtins.str]] = None,
                 name: Optional[pulumi.Input[builtins.str]] = None,
                 owner: Optional[pulumi.Input[builtins.str]] = None,
                 state: Optional[pulumi.Input[builtins.str]] = None,
                 type: Optional[pulumi.Input[builtins.str]] = None,
                 updated_at: Optional[pulumi.Input[builtins.str]] = None):
        """
        Input properties used for looking up and filtering SchemaRegistry resources.
        :param pulumi.Input['SchemaRegistryConfluentArgs'] confluent: Confluent specific configuration
        :param pulumi.Input['SchemaRegistryConfluentCloudArgs'] confluent_cloud: Confluent cloud specific configuration
        :param pulumi.Input[builtins.str] created_at: Creation date of the schema registry
        :param pulumi.Input[builtins.str] name: Name of the schema registry
        :param pulumi.Input[builtins.str] owner: Owning role of the schema registry
        :param pulumi.Input[builtins.str] state: Status of the schema registry
        :param pulumi.Input[builtins.str] type: Type of the schema registry
        :param pulumi.Input[builtins.str] updated_at: Last update date of the schema registry
        """
        if confluent is not None:
            pulumi.set(__self__, "confluent", confluent)
        if confluent_cloud is not None:
            pulumi.set(__self__, "confluent_cloud", confluent_cloud)
        if created_at is not None:
            pulumi.set(__self__, "created_at", created_at)
        if name is not None:
            pulumi.set(__self__, "name", name)
        if owner is not None:
            pulumi.set(__self__, "owner", owner)
        if state is not None:
            pulumi.set(__self__, "state", state)
        if type is not None:
            pulumi.set(__self__, "type", type)
        if updated_at is not None:
            pulumi.set(__self__, "updated_at", updated_at)

    @property
    @pulumi.getter
    def confluent(self) -> Optional[pulumi.Input['SchemaRegistryConfluentArgs']]:
        """
        Confluent specific configuration
        """
        return pulumi.get(self, "confluent")

    @confluent.setter
    def confluent(self, value: Optional[pulumi.Input['SchemaRegistryConfluentArgs']]):
        pulumi.set(self, "confluent", value)

    @property
    @pulumi.getter(name="confluentCloud")
    def confluent_cloud(self) -> Optional[pulumi.Input['SchemaRegistryConfluentCloudArgs']]:
        """
        Confluent cloud specific configuration
        """
        return pulumi.get(self, "confluent_cloud")

    @confluent_cloud.setter
    def confluent_cloud(self, value: Optional[pulumi.Input['SchemaRegistryConfluentCloudArgs']]):
        pulumi.set(self, "confluent_cloud", value)

    @property
    @pulumi.getter(name="createdAt")
    def created_at(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Creation date of the schema registry
        """
        return pulumi.get(self, "created_at")

    @created_at.setter
    def created_at(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "created_at", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Name of the schema registry
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter
    def owner(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Owning role of the schema registry
        """
        return pulumi.get(self, "owner")

    @owner.setter
    def owner(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "owner", value)

    @property
    @pulumi.getter
    def state(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Status of the schema registry
        """
        return pulumi.get(self, "state")

    @state.setter
    def state(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "state", value)

    @property
    @pulumi.getter
    def type(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Type of the schema registry
        """
        return pulumi.get(self, "type")

    @type.setter
    def type(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "type", value)

    @property
    @pulumi.getter(name="updatedAt")
    def updated_at(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Last update date of the schema registry
        """
        return pulumi.get(self, "updated_at")

    @updated_at.setter
    def updated_at(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "updated_at", value)


@pulumi.type_token("deltastream:index/schemaRegistry:SchemaRegistry")
class SchemaRegistry(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 confluent: Optional[pulumi.Input[Union['SchemaRegistryConfluentArgs', 'SchemaRegistryConfluentArgsDict']]] = None,
                 confluent_cloud: Optional[pulumi.Input[Union['SchemaRegistryConfluentCloudArgs', 'SchemaRegistryConfluentCloudArgsDict']]] = None,
                 name: Optional[pulumi.Input[builtins.str]] = None,
                 owner: Optional[pulumi.Input[builtins.str]] = None,
                 __props__=None):
        """
        Schema registry resource

        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[Union['SchemaRegistryConfluentArgs', 'SchemaRegistryConfluentArgsDict']] confluent: Confluent specific configuration
        :param pulumi.Input[Union['SchemaRegistryConfluentCloudArgs', 'SchemaRegistryConfluentCloudArgsDict']] confluent_cloud: Confluent cloud specific configuration
        :param pulumi.Input[builtins.str] name: Name of the schema registry
        :param pulumi.Input[builtins.str] owner: Owning role of the schema registry
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: Optional[SchemaRegistryArgs] = None,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Schema registry resource

        :param str resource_name: The name of the resource.
        :param SchemaRegistryArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(SchemaRegistryArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 confluent: Optional[pulumi.Input[Union['SchemaRegistryConfluentArgs', 'SchemaRegistryConfluentArgsDict']]] = None,
                 confluent_cloud: Optional[pulumi.Input[Union['SchemaRegistryConfluentCloudArgs', 'SchemaRegistryConfluentCloudArgsDict']]] = None,
                 name: Optional[pulumi.Input[builtins.str]] = None,
                 owner: Optional[pulumi.Input[builtins.str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = SchemaRegistryArgs.__new__(SchemaRegistryArgs)

            __props__.__dict__["confluent"] = confluent
            __props__.__dict__["confluent_cloud"] = confluent_cloud
            __props__.__dict__["name"] = name
            __props__.__dict__["owner"] = owner
            __props__.__dict__["created_at"] = None
            __props__.__dict__["state"] = None
            __props__.__dict__["type"] = None
            __props__.__dict__["updated_at"] = None
        super(SchemaRegistry, __self__).__init__(
            'deltastream:index/schemaRegistry:SchemaRegistry',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None,
            confluent: Optional[pulumi.Input[Union['SchemaRegistryConfluentArgs', 'SchemaRegistryConfluentArgsDict']]] = None,
            confluent_cloud: Optional[pulumi.Input[Union['SchemaRegistryConfluentCloudArgs', 'SchemaRegistryConfluentCloudArgsDict']]] = None,
            created_at: Optional[pulumi.Input[builtins.str]] = None,
            name: Optional[pulumi.Input[builtins.str]] = None,
            owner: Optional[pulumi.Input[builtins.str]] = None,
            state: Optional[pulumi.Input[builtins.str]] = None,
            type: Optional[pulumi.Input[builtins.str]] = None,
            updated_at: Optional[pulumi.Input[builtins.str]] = None) -> 'SchemaRegistry':
        """
        Get an existing SchemaRegistry resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[Union['SchemaRegistryConfluentArgs', 'SchemaRegistryConfluentArgsDict']] confluent: Confluent specific configuration
        :param pulumi.Input[Union['SchemaRegistryConfluentCloudArgs', 'SchemaRegistryConfluentCloudArgsDict']] confluent_cloud: Confluent cloud specific configuration
        :param pulumi.Input[builtins.str] created_at: Creation date of the schema registry
        :param pulumi.Input[builtins.str] name: Name of the schema registry
        :param pulumi.Input[builtins.str] owner: Owning role of the schema registry
        :param pulumi.Input[builtins.str] state: Status of the schema registry
        :param pulumi.Input[builtins.str] type: Type of the schema registry
        :param pulumi.Input[builtins.str] updated_at: Last update date of the schema registry
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = _SchemaRegistryState.__new__(_SchemaRegistryState)

        __props__.__dict__["confluent"] = confluent
        __props__.__dict__["confluent_cloud"] = confluent_cloud
        __props__.__dict__["created_at"] = created_at
        __props__.__dict__["name"] = name
        __props__.__dict__["owner"] = owner
        __props__.__dict__["state"] = state
        __props__.__dict__["type"] = type
        __props__.__dict__["updated_at"] = updated_at
        return SchemaRegistry(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter
    def confluent(self) -> pulumi.Output[Optional['outputs.SchemaRegistryConfluent']]:
        """
        Confluent specific configuration
        """
        return pulumi.get(self, "confluent")

    @property
    @pulumi.getter(name="confluentCloud")
    def confluent_cloud(self) -> pulumi.Output[Optional['outputs.SchemaRegistryConfluentCloud']]:
        """
        Confluent cloud specific configuration
        """
        return pulumi.get(self, "confluent_cloud")

    @property
    @pulumi.getter(name="createdAt")
    def created_at(self) -> pulumi.Output[builtins.str]:
        """
        Creation date of the schema registry
        """
        return pulumi.get(self, "created_at")

    @property
    @pulumi.getter
    def name(self) -> pulumi.Output[builtins.str]:
        """
        Name of the schema registry
        """
        return pulumi.get(self, "name")

    @property
    @pulumi.getter
    def owner(self) -> pulumi.Output[builtins.str]:
        """
        Owning role of the schema registry
        """
        return pulumi.get(self, "owner")

    @property
    @pulumi.getter
    def state(self) -> pulumi.Output[builtins.str]:
        """
        Status of the schema registry
        """
        return pulumi.get(self, "state")

    @property
    @pulumi.getter
    def type(self) -> pulumi.Output[builtins.str]:
        """
        Type of the schema registry
        """
        return pulumi.get(self, "type")

    @property
    @pulumi.getter(name="updatedAt")
    def updated_at(self) -> pulumi.Output[builtins.str]:
        """
        Last update date of the schema registry
        """
        return pulumi.get(self, "updated_at")

