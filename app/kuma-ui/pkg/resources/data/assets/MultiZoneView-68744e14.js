import{L as x}from"./LoadingBox-04e11492.js";import{O as z,a as Z,b as I}from"./OnboardingPage-a5993aa7.js";import{u as T,a as k,g as A,f as O,e as N,_ as V}from"./RouteView.vue_vue_type_script_setup_true_lang-49cffe7a.js";import{_ as L}from"./RouteTitle.vue_vue_type_script_setup_true_lang-994a27b4.js";import{d as M,r,i as U,o as a,c as C,w as c,a as i,u as m,b as e,e as n,f as l,g as S,p as B,j as $}from"./index-2aa994fe.js";import"./kongponents.es-7e228e6a.js";const E=u=>(B("data-v-40900992"),u=u(),$(),u),K=E(()=>n("p",{class:"mb-4 text-center"},`
            A zone requires both the zone control plane and zone ingress. On Kubernetes, you run a single command to create both resources. On Universal, you must create them separately.
          `,-1)),R={class:"mb-4 text-center"},D=["href"],P={class:"status-box mt-4"},G={key:0,class:"status--is-connected","data-testid":"zone-connected"},j={key:1,class:"status--is-disconnected","data-testid":"zone-disconnected"},q={class:"status-box mt-4"},H={key:0,class:"status--is-connected","data-testid":"zone-ingress-connected"},Q={key:1,class:"status--is-disconnected","data-testid":"zone-ingress-disconnected"},Y={key:0,class:"status-loading-box mt-4"},b=1e3,F=M({__name:"MultiZoneView",setup(u){const p=T(),f=k(),{t:y}=A(),s=r(!1),o=r(!1),d=r(null),_=r(null);U(function(){h(),w()}),v(),g();async function v(){try{const{total:t}=await p.getZones();s.value=t>0}catch(t){s.value=!1,console.error(t)}finally{s.value||(h(),d.value=window.setTimeout(v,b))}}async function g(){try{const{total:t}=await p.getAllZoneIngressOverviews();o.value=t>0}catch(t){o.value=!1,console.error(t)}finally{o.value||(w(),_.value=window.setTimeout(g,b))}}function h(){d.value!==null&&window.clearTimeout(d.value)}function w(){_.value!==null&&window.clearTimeout(_.value)}return(t,J)=>(a(),C(N,null,{default:c(()=>[i(L,{title:m(y)("onboarding.routes.multizone.title")},null,8,["title"]),e(),i(O,null,{default:c(()=>[i(z,null,{header:c(()=>[i(Z,null,{title:c(()=>[e(`
              Add zones
            `)]),_:1})]),content:c(()=>[K,e(),n("p",R,[n("b",null,[e("See "),n("a",{href:`${m(f)("KUMA_DOCS_URL")}/deployments/multi-zone/?${m(f)("KUMA_UTM_QUERY_PARAMS")}#zone-control-plane`,target:"_blank"},"the documentation for options to install",8,D),e(".")])]),e(),n("div",null,[n("p",P,[e(`
              Zone status:

              `),s.value?(a(),l("span",G,"Connected")):(a(),l("span",j,"Disconnected"))]),e(),n("p",q,[e(`
              Zone ingress status:

              `),o.value?(a(),l("span",H,"Connected")):(a(),l("span",Q,"Disconnected"))]),e(),!o.value||!s.value?(a(),l("div",Y,[i(x)])):S("",!0)])]),navigation:c(()=>[i(I,{"next-step":"onboarding-create-mesh","previous-step":"onboarding-configuration-types","should-allow-next":s.value&&o.value},null,8,["should-allow-next"])]),_:1})]),_:1})]),_:1}))}});const oe=V(F,[["__scopeId","data-v-40900992"]]);export{oe as default};
