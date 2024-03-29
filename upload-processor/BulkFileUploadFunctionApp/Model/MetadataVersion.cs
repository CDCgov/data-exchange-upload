using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Model
{
    public abstract class Enumeration
    {
        public int Value { get; }

        protected Enumeration(int value)
        {
            Value = value;
        }
    }
    public class MetadataVersion : Enumeration
    {
        public static readonly MetadataVersion V1 = new MetadataVersion(1);
        public static readonly MetadataVersion V2 = new MetadataVersion(2);
        public MetadataVersion(int value): base(value) { }

        public override string ToString()
        {
            switch(Value)
            {
                case 1 when this == V1:
                    return "1.0";
                case 2 when this == V2:
                    return "2.0";
                default:
                    throw new InvalidOperationException($"Unsupported value: {Value}");
            }
        }

        public static MetadataVersion FromString(string value)
        {
            switch(value)
            {
                case "1.0":
                    return V1;
                case "2.0":
                    return V2;
                default:
                    throw new ArgumentException($"{value} is not supported version.");
            }
        }
    }
}
