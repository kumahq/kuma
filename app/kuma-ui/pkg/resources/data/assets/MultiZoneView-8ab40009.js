import{L as z}from"./LoadingBox-5ca41aec.js";import{O as T,a as V,b as k}from"./OnboardingPage-ac38a003.js";import{d as I,P as O,v as r,Q as A,r as p,o as a,g as N,w as i,h as c,l as e,m as n,j as l,k as C,B,C as L,q as R}from"./index-ecc7df9d.js";const S=u=>(B("data-v-052795d6"),u=u(),L(),u),M=S(()=>n("p",{class:"mb-4 text-center"},`
            A zone requires both the zone control plane and zone ingress. On Kubernetes, you run a single command to create both resources. On Universal, you must create them separately.
          `,-1)),P={class:"mb-4 text-center"},q=["href"],D={class:"status-box mt-4"},E={key:0,class:"status--is-connected","data-testid":"zone-connected"},G={key:1,class:"status--is-disconnected","data-testid":"zone-disconnected"},K={class:"status-box mt-4"},U={key:0,class:"status--is-connected","data-testid":"zone-ingress-connected"},j={key:1,class:"status--is-disconnected","data-testid":"zone-ingress-disconnected"},H={key:0,class:"status-loading-box mt-4"},b=1e3,Q=I({__name:"MultiZoneView",setup(u){const m=O(),o=r(!1),s=r(!1),d=r(null),_=r(null);A(function(){h(),f()}),g(),v();async function g(){try{const{total:t}=await m.getZones();o.value=t>0}catch(t){o.value=!1,console.error(t)}finally{o.value||(h(),d.value=window.setTimeout(g,b))}}async function v(){try{const{total:t}=await m.getAllZoneIngressOverviews();s.value=t>0}catch(t){s.value=!1,console.error(t)}finally{s.value||(f(),_.value=window.setTimeout(v,b))}}function h(){d.value!==null&&window.clearTimeout(d.value)}function f(){_.value!==null&&window.clearTimeout(_.value)}return(t,F)=>{const y=p("RouteTitle"),x=p("AppView"),Z=p("RouteView");return a(),N(Z,{name:"onboarding-multi-zone"},{default:i(({t:w})=>[c(y,{title:w("onboarding.routes.multizone.title")},null,8,["title"]),e(),c(x,null,{default:i(()=>[c(T,null,{header:i(()=>[c(V,null,{title:i(()=>[e(`
              Add zones
            `)]),_:1})]),content:i(()=>[M,e(),n("p",P,[n("b",null,[e("See "),n("a",{href:w("onboarding.href.docs.install"),target:"_blank"},"the documentation for options to install",8,q),e(".")])]),e(),n("div",null,[n("p",D,[e(`
              Zone status:

              `),o.value?(a(),l("span",E,"Connected")):(a(),l("span",G,"Disconnected"))]),e(),n("p",K,[e(`
              Zone ingress status:

              `),s.value?(a(),l("span",U,"Connected")):(a(),l("span",j,"Disconnected"))]),e(),!s.value||!o.value?(a(),l("div",H,[c(z)])):C("",!0)])]),navigation:i(()=>[c(k,{"next-step":"onboarding-create-mesh","previous-step":"onboarding-configuration-types","should-allow-next":o.value&&s.value},null,8,["should-allow-next"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const Y=R(Q,[["__scopeId","data-v-052795d6"]]);export{Y as default};
