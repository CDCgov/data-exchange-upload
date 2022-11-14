using System;
using System.Collections.Generic;
using System.Text;

namespace TUSTestHelpers.Behaviors
{
    public static class ServerExpectedBehaviors {
        public static readonly string ServerExpectedBehaviorKey = nameof(ServerExpectedBehaviors);
        
        public static (string k, string v)[] Default = new (string k, string v)[] { (nameof(ServerExpectedBehaviors.Default), nameof(ServerExpectedBehaviors.Default)) };
        public static (string k, string v)[] ForceChunkFailure = new (string k, string v)[] { (nameof(ServerExpectedBehaviors.ForceChunkFailure), nameof(ServerExpectedBehaviors.ForceChunkFailure)) };
        public static (string k, string v)[] ForceFullFailure = new (string k, string v)[] { (nameof(ServerExpectedBehaviors.ForceFullFailure), nameof(ServerExpectedBehaviors.ForceFullFailure)) };


        //{
        //new Tuple<string, string>(nameof(ServerExpectedBehavior.ForceChunkFailure), nameof(ServerExpectedBehavior.ForceChunkFailure))
        //};
    }
    public enum ServerExpectedBehavior
    {
        Default = 0,
        ForceChunkFailure = 1,
        ForceFullFailure = 2
    }
}
