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

__all__ = ['SecretArgs', 'Secret']

@pulumi.input_type
class SecretArgs:
    def __init__(__self__, *,
                 type: pulumi.Input[builtins.str],
                 custom_properties: Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]] = None,
                 description: Optional[pulumi.Input[builtins.str]] = None,
                 name: Optional[pulumi.Input[builtins.str]] = None,
                 owner: Optional[pulumi.Input[builtins.str]] = None,
                 string_value: Optional[pulumi.Input[builtins.str]] = None):
        """
        The set of arguments for constructing a Secret resource.
        :param pulumi.Input[builtins.str] type: Secret type. (Valid values: generic_string)
        :param pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]] custom_properties: Custom properties of the Secret
        :param pulumi.Input[builtins.str] description: Description of the Secret
        :param pulumi.Input[builtins.str] name: Name of the Secret
        :param pulumi.Input[builtins.str] owner: Owning role of the Secret
        :param pulumi.Input[builtins.str] string_value: Secret value
        """
        pulumi.set(__self__, "type", type)
        if custom_properties is not None:
            pulumi.set(__self__, "custom_properties", custom_properties)
        if description is not None:
            pulumi.set(__self__, "description", description)
        if name is not None:
            pulumi.set(__self__, "name", name)
        if owner is not None:
            pulumi.set(__self__, "owner", owner)
        if string_value is not None:
            pulumi.set(__self__, "string_value", string_value)

    @property
    @pulumi.getter
    def type(self) -> pulumi.Input[builtins.str]:
        """
        Secret type. (Valid values: generic_string)
        """
        return pulumi.get(self, "type")

    @type.setter
    def type(self, value: pulumi.Input[builtins.str]):
        pulumi.set(self, "type", value)

    @property
    @pulumi.getter(name="customProperties")
    def custom_properties(self) -> Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]]:
        """
        Custom properties of the Secret
        """
        return pulumi.get(self, "custom_properties")

    @custom_properties.setter
    def custom_properties(self, value: Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]]):
        pulumi.set(self, "custom_properties", value)

    @property
    @pulumi.getter
    def description(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Description of the Secret
        """
        return pulumi.get(self, "description")

    @description.setter
    def description(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "description", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Name of the Secret
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter
    def owner(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Owning role of the Secret
        """
        return pulumi.get(self, "owner")

    @owner.setter
    def owner(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "owner", value)

    @property
    @pulumi.getter(name="stringValue")
    def string_value(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Secret value
        """
        return pulumi.get(self, "string_value")

    @string_value.setter
    def string_value(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "string_value", value)


@pulumi.input_type
class _SecretState:
    def __init__(__self__, *,
                 created_at: Optional[pulumi.Input[builtins.str]] = None,
                 custom_properties: Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]] = None,
                 description: Optional[pulumi.Input[builtins.str]] = None,
                 name: Optional[pulumi.Input[builtins.str]] = None,
                 owner: Optional[pulumi.Input[builtins.str]] = None,
                 status: Optional[pulumi.Input[builtins.str]] = None,
                 string_value: Optional[pulumi.Input[builtins.str]] = None,
                 type: Optional[pulumi.Input[builtins.str]] = None,
                 updated_at: Optional[pulumi.Input[builtins.str]] = None):
        """
        Input properties used for looking up and filtering Secret resources.
        :param pulumi.Input[builtins.str] created_at: Creation date of the Secret
        :param pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]] custom_properties: Custom properties of the Secret
        :param pulumi.Input[builtins.str] description: Description of the Secret
        :param pulumi.Input[builtins.str] name: Name of the Secret
        :param pulumi.Input[builtins.str] owner: Owning role of the Secret
        :param pulumi.Input[builtins.str] status: Status of the Secret
        :param pulumi.Input[builtins.str] string_value: Secret value
        :param pulumi.Input[builtins.str] type: Secret type. (Valid values: generic_string)
        :param pulumi.Input[builtins.str] updated_at: Last update date of the Secret
        """
        if created_at is not None:
            pulumi.set(__self__, "created_at", created_at)
        if custom_properties is not None:
            pulumi.set(__self__, "custom_properties", custom_properties)
        if description is not None:
            pulumi.set(__self__, "description", description)
        if name is not None:
            pulumi.set(__self__, "name", name)
        if owner is not None:
            pulumi.set(__self__, "owner", owner)
        if status is not None:
            pulumi.set(__self__, "status", status)
        if string_value is not None:
            pulumi.set(__self__, "string_value", string_value)
        if type is not None:
            pulumi.set(__self__, "type", type)
        if updated_at is not None:
            pulumi.set(__self__, "updated_at", updated_at)

    @property
    @pulumi.getter(name="createdAt")
    def created_at(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Creation date of the Secret
        """
        return pulumi.get(self, "created_at")

    @created_at.setter
    def created_at(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "created_at", value)

    @property
    @pulumi.getter(name="customProperties")
    def custom_properties(self) -> Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]]:
        """
        Custom properties of the Secret
        """
        return pulumi.get(self, "custom_properties")

    @custom_properties.setter
    def custom_properties(self, value: Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]]):
        pulumi.set(self, "custom_properties", value)

    @property
    @pulumi.getter
    def description(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Description of the Secret
        """
        return pulumi.get(self, "description")

    @description.setter
    def description(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "description", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Name of the Secret
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter
    def owner(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Owning role of the Secret
        """
        return pulumi.get(self, "owner")

    @owner.setter
    def owner(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "owner", value)

    @property
    @pulumi.getter
    def status(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Status of the Secret
        """
        return pulumi.get(self, "status")

    @status.setter
    def status(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "status", value)

    @property
    @pulumi.getter(name="stringValue")
    def string_value(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Secret value
        """
        return pulumi.get(self, "string_value")

    @string_value.setter
    def string_value(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "string_value", value)

    @property
    @pulumi.getter
    def type(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Secret type. (Valid values: generic_string)
        """
        return pulumi.get(self, "type")

    @type.setter
    def type(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "type", value)

    @property
    @pulumi.getter(name="updatedAt")
    def updated_at(self) -> Optional[pulumi.Input[builtins.str]]:
        """
        Last update date of the Secret
        """
        return pulumi.get(self, "updated_at")

    @updated_at.setter
    def updated_at(self, value: Optional[pulumi.Input[builtins.str]]):
        pulumi.set(self, "updated_at", value)


@pulumi.type_token("deltastream:index/secret:Secret")
class Secret(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 custom_properties: Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]] = None,
                 description: Optional[pulumi.Input[builtins.str]] = None,
                 name: Optional[pulumi.Input[builtins.str]] = None,
                 owner: Optional[pulumi.Input[builtins.str]] = None,
                 string_value: Optional[pulumi.Input[builtins.str]] = None,
                 type: Optional[pulumi.Input[builtins.str]] = None,
                 __props__=None):
        """
        Secret resource

        ## Example Usage

        ```python
        import pulumi
        import deltastream-pulumi as deltastream

        example = deltastream.Secret("example",
            name="example_secret",
            type="generic_string",
            description="secret description",
            string_value="some value")
        ```

        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]] custom_properties: Custom properties of the Secret
        :param pulumi.Input[builtins.str] description: Description of the Secret
        :param pulumi.Input[builtins.str] name: Name of the Secret
        :param pulumi.Input[builtins.str] owner: Owning role of the Secret
        :param pulumi.Input[builtins.str] string_value: Secret value
        :param pulumi.Input[builtins.str] type: Secret type. (Valid values: generic_string)
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: SecretArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Secret resource

        ## Example Usage

        ```python
        import pulumi
        import deltastream-pulumi as deltastream

        example = deltastream.Secret("example",
            name="example_secret",
            type="generic_string",
            description="secret description",
            string_value="some value")
        ```

        :param str resource_name: The name of the resource.
        :param SecretArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(SecretArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 custom_properties: Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]] = None,
                 description: Optional[pulumi.Input[builtins.str]] = None,
                 name: Optional[pulumi.Input[builtins.str]] = None,
                 owner: Optional[pulumi.Input[builtins.str]] = None,
                 string_value: Optional[pulumi.Input[builtins.str]] = None,
                 type: Optional[pulumi.Input[builtins.str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = SecretArgs.__new__(SecretArgs)

            __props__.__dict__["custom_properties"] = custom_properties
            __props__.__dict__["description"] = description
            __props__.__dict__["name"] = name
            __props__.__dict__["owner"] = owner
            __props__.__dict__["string_value"] = string_value
            if type is None and not opts.urn:
                raise TypeError("Missing required property 'type'")
            __props__.__dict__["type"] = type
            __props__.__dict__["created_at"] = None
            __props__.__dict__["status"] = None
            __props__.__dict__["updated_at"] = None
        super(Secret, __self__).__init__(
            'deltastream:index/secret:Secret',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None,
            created_at: Optional[pulumi.Input[builtins.str]] = None,
            custom_properties: Optional[pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]]] = None,
            description: Optional[pulumi.Input[builtins.str]] = None,
            name: Optional[pulumi.Input[builtins.str]] = None,
            owner: Optional[pulumi.Input[builtins.str]] = None,
            status: Optional[pulumi.Input[builtins.str]] = None,
            string_value: Optional[pulumi.Input[builtins.str]] = None,
            type: Optional[pulumi.Input[builtins.str]] = None,
            updated_at: Optional[pulumi.Input[builtins.str]] = None) -> 'Secret':
        """
        Get an existing Secret resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[builtins.str] created_at: Creation date of the Secret
        :param pulumi.Input[Mapping[str, pulumi.Input[builtins.str]]] custom_properties: Custom properties of the Secret
        :param pulumi.Input[builtins.str] description: Description of the Secret
        :param pulumi.Input[builtins.str] name: Name of the Secret
        :param pulumi.Input[builtins.str] owner: Owning role of the Secret
        :param pulumi.Input[builtins.str] status: Status of the Secret
        :param pulumi.Input[builtins.str] string_value: Secret value
        :param pulumi.Input[builtins.str] type: Secret type. (Valid values: generic_string)
        :param pulumi.Input[builtins.str] updated_at: Last update date of the Secret
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = _SecretState.__new__(_SecretState)

        __props__.__dict__["created_at"] = created_at
        __props__.__dict__["custom_properties"] = custom_properties
        __props__.__dict__["description"] = description
        __props__.__dict__["name"] = name
        __props__.__dict__["owner"] = owner
        __props__.__dict__["status"] = status
        __props__.__dict__["string_value"] = string_value
        __props__.__dict__["type"] = type
        __props__.__dict__["updated_at"] = updated_at
        return Secret(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter(name="createdAt")
    def created_at(self) -> pulumi.Output[builtins.str]:
        """
        Creation date of the Secret
        """
        return pulumi.get(self, "created_at")

    @property
    @pulumi.getter(name="customProperties")
    def custom_properties(self) -> pulumi.Output[Optional[Mapping[str, builtins.str]]]:
        """
        Custom properties of the Secret
        """
        return pulumi.get(self, "custom_properties")

    @property
    @pulumi.getter
    def description(self) -> pulumi.Output[Optional[builtins.str]]:
        """
        Description of the Secret
        """
        return pulumi.get(self, "description")

    @property
    @pulumi.getter
    def name(self) -> pulumi.Output[builtins.str]:
        """
        Name of the Secret
        """
        return pulumi.get(self, "name")

    @property
    @pulumi.getter
    def owner(self) -> pulumi.Output[builtins.str]:
        """
        Owning role of the Secret
        """
        return pulumi.get(self, "owner")

    @property
    @pulumi.getter
    def status(self) -> pulumi.Output[builtins.str]:
        """
        Status of the Secret
        """
        return pulumi.get(self, "status")

    @property
    @pulumi.getter(name="stringValue")
    def string_value(self) -> pulumi.Output[Optional[builtins.str]]:
        """
        Secret value
        """
        return pulumi.get(self, "string_value")

    @property
    @pulumi.getter
    def type(self) -> pulumi.Output[builtins.str]:
        """
        Secret type. (Valid values: generic_string)
        """
        return pulumi.get(self, "type")

    @property
    @pulumi.getter(name="updatedAt")
    def updated_at(self) -> pulumi.Output[builtins.str]:
        """
        Last update date of the Secret
        """
        return pulumi.get(self, "updated_at")

