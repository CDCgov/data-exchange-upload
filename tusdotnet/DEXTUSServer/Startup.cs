using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.HttpsPolicy;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using tusdotnet;
using tusdotnet.Interfaces;
using tusdotnet.Models;
using tusdotnet.Models.Configuration;
using tusdotnet.Stores;
using TUSTestHelpers.Behaviors;
using TUSTestHelpers.Behaviors.Server;

namespace DEXTUSServer
{
    public class Startup
    {
        public Startup(IConfiguration configuration)
        {
            Configuration = configuration;
        }

        public IConfiguration Configuration { get; }

        // This method gets called by the runtime. Use this method to add services to the container.
        public void ConfigureServices(IServiceCollection services)
        {
            services.AddRazorPages();
        }

        // This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
        public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
            }
            else
            {
                app.UseExceptionHandler("/Error");
                // The default HSTS value is 30 days. You may want to change this for production scenarios, see https://aka.ms/aspnetcore-hsts.
                app.UseHsts();
            }

            app.UseHttpsRedirection();
            app.UseStaticFiles();

            app.UseRouting();

            app.UseAuthorization();


            app.UseTus(httpContext => new DefaultTusConfiguration
            {
                // This method is called on each request so different configurations can be returned per user, domain, path etc.
                // Return null to disable tusdotnet for the current request.

                // c:\tusfiles is where to store files
                Store = new TusDiskStore(@"C:\filestore\uploads\"),
                // On what url should we listen for uploads?
                UrlPath = "/files",
                Events = new Events
                {
                    
                    OnFileCompleteAsync = async eventContext =>
                    {
                        ITusFile file = await eventContext.GetFileAsync();
                        var fileId = file.Id;
                        Dictionary<string, Metadata> metadata = await file.GetMetadataAsync(eventContext.CancellationToken);
                        var behavior = metadata.FirstOrDefault();

                        //Example Serverside Reactors
                        switch (behavior.Key) {

                            case "blah":
                                break;
                            case "skip" + nameof(ServerExpectedBehavior.ForceChunkFailure):
                                break;
                            default:
                                await DefaultServer.ProcessPostUpload(file, eventContext);
                                break;
                        }
                       
                    }
                }
            });


            app.UseEndpoints(endpoints =>
            {
                endpoints.MapRazorPages();
            });
        }
    }
}
