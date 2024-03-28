
public class UriWrapper : IUriWrapper
{
    private readonly Uri _uri;

    public UriWrapper(string uriString)
    {
        _uri = new Uri(uriString);
    }

    public Uri GetUri()
    {
        return _uri;
    }
}