import{O as c,a as g,b as v}from"./OnboardingPage-BCkL46S4.js";import{d as f,v as _,r as a,o as w,p as y,w as t,b as n,e as o,l as s,t as x,_ as V}from"./index-yoi81zLz.js";const A={class:"mb-4 text-center"},T=f({__name:"OnboardingCreateMeshView",setup(C){const r=[{label:"Name",key:"name"},{label:"Services",key:"servicesAmount"},{label:"DPPs",key:"dppsAmount"}],l=_({total:1,data:[{name:"default",servicesAmount:0,dppsAmount:0}]});return(N,e)=>{const d=a("RouteTitle"),p=a("KTable"),u=a("AppView"),m=a("RouteView");return w(),y(m,{name:"onboarding-create-mesh-view"},{default:t(({can:b,t:i})=>[n(d,{title:i("onboarding.routes.create-mesh.title"),render:!1},null,8,["title"]),e[8]||(e[8]=o()),n(u,null,{default:t(()=>[n(c,null,{header:t(()=>[n(g,null,{title:t(()=>e[0]||(e[0]=[o(`
              Create the mesh
            `)])),_:1})]),content:t(()=>[s("p",A,[o(`
            When you install, `+x(i("common.product.name"))+" creates a ",1),e[1]||(e[1]=s("i",null,"default",-1)),e[2]||(e[2]=o(` mesh, but you can add as many meshes as you need.
          `))]),e[3]||(e[3]=o()),n(p,{class:"table",fetcher:()=>l.value,headers:r,"disable-pagination":""},null,8,["fetcher"]),e[4]||(e[4]=o()),e[5]||(e[5]=s("p",{class:"mt-4 text-center"},`
            This mesh is empty. Next, you add services and their data plane proxies.
          `,-1))]),navigation:t(()=>[n(v,{"next-step":"onboarding-add-new-services-view","previous-step":b("use zones")?"onboarding-multi-zone-view":"onboarding-configuration-types-view"},null,8,["previous-step"])]),_:2},1024)]),_:2},1024)]),_:1})}}}),k=V(T,[["__scopeId","data-v-24e81496"]]);export{k as default};
