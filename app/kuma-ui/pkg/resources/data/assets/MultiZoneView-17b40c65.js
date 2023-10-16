import{L as z}from"./LoadingBox-549a5c77.js";import{O as T,a as V,b as I}from"./OnboardingPage-9bf33915.js";import{d as k,R as O,y as r,S as A,r as p,o as a,i as N,w as i,j as c,n as e,p as n,l,m as C,D as L,G as R,t as B}from"./index-21079cd9.js";const S=u=>(L("data-v-052795d6"),u=u(),R(),u),D=S(()=>n("p",{class:"mb-4 text-center"},`
            A zone requires both the zone control plane and zone ingress. On Kubernetes, you run a single command to create both resources. On Universal, you must create them separately.
          `,-1)),G={class:"mb-4 text-center"},M=["href"],E={class:"status-box mt-4"},K={key:0,class:"status--is-connected","data-testid":"zone-connected"},P={key:1,class:"status--is-disconnected","data-testid":"zone-disconnected"},U={class:"status-box mt-4"},j={key:0,class:"status--is-connected","data-testid":"zone-ingress-connected"},q={key:1,class:"status--is-disconnected","data-testid":"zone-ingress-disconnected"},H={key:0,class:"status-loading-box mt-4"},b=1e3,F=k({__name:"MultiZoneView",setup(u){const m=O(),o=r(!1),s=r(!1),d=r(null),_=r(null);A(function(){f(),h()}),g(),v();async function g(){try{const{total:t}=await m.getZones();o.value=t>0}catch(t){o.value=!1,console.error(t)}finally{o.value||(f(),d.value=window.setTimeout(g,b))}}async function v(){try{const{total:t}=await m.getAllZoneIngressOverviews();s.value=t>0}catch(t){s.value=!1,console.error(t)}finally{s.value||(h(),_.value=window.setTimeout(v,b))}}function f(){d.value!==null&&window.clearTimeout(d.value)}function h(){_.value!==null&&window.clearTimeout(_.value)}return(t,J)=>{const y=p("RouteTitle"),x=p("AppView"),Z=p("RouteView");return a(),N(Z,{name:"onboarding-multi-zone"},{default:i(({t:w})=>[c(y,{title:w("onboarding.routes.multizone.title")},null,8,["title"]),e(),c(x,null,{default:i(()=>[c(T,null,{header:i(()=>[c(V,null,{title:i(()=>[e(`
              Add zones
            `)]),_:1})]),content:i(()=>[D,e(),n("p",G,[n("b",null,[e("See "),n("a",{href:w("onboarding.href.docs.install"),target:"_blank"},"the documentation for options to install",8,M),e(".")])]),e(),n("div",null,[n("p",E,[e(`
              Zone status:

              `),o.value?(a(),l("span",K,"Connected")):(a(),l("span",P,"Disconnected"))]),e(),n("p",U,[e(`
              Zone ingress status:

              `),s.value?(a(),l("span",j,"Connected")):(a(),l("span",q,"Disconnected"))]),e(),!s.value||!o.value?(a(),l("div",H,[c(z)])):C("",!0)])]),navigation:i(()=>[c(I,{"next-step":"onboarding-create-mesh","previous-step":"onboarding-configuration-types","should-allow-next":o.value&&s.value},null,8,["should-allow-next"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const Y=B(F,[["__scopeId","data-v-052795d6"]]);export{Y as default};
