import{L as y}from"./LoadingBox-16c3056e.js";import{O as x,a as Z,b as z}from"./OnboardingPage-d7f0da66.js";import{l as I,h as k,A as T,_ as O,f as V}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";import{_ as A}from"./RouteTitle.vue_vue_type_script_setup_true_lang-6484968f.js";import{d as N,B as u,a3 as B,o as a,a as L,w as i,h as c,b as w,g as e,i as n,e as l,f as C,y as S,z as M}from"./index-f1b8ae6a.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-14dd845b.js";const D=r=>(S("data-v-5d5ad5e8"),r=r(),M(),r),E=D(()=>n("p",{class:"mb-4 text-center"},`
            A zone requires both the zone control plane and zone ingress. On Kubernetes, you run a single command to create both resources. On Universal, you must create them separately.
          `,-1)),G={class:"mb-4 text-center"},K=["href"],P={class:"status-box mt-4"},U={key:0,class:"status--is-connected","data-testid":"zone-connected"},$={key:1,class:"status--is-disconnected","data-testid":"zone-disconnected"},q={class:"status-box mt-4"},H={key:0,class:"status--is-connected","data-testid":"zone-ingress-connected"},R={key:1,class:"status--is-disconnected","data-testid":"zone-ingress-disconnected"},j={key:0,class:"status-loading-box mt-4"},b=1e3,F=N({__name:"MultiZoneView",setup(r){const m=I(),{t:p}=k(),s=u(!1),o=u(!1),d=u(null),_=u(null);B(function(){g(),v()}),f(),h();async function f(){try{const{total:t}=await m.getZones();s.value=t>0}catch(t){s.value=!1,console.error(t)}finally{s.value||(g(),d.value=window.setTimeout(f,b))}}async function h(){try{const{total:t}=await m.getAllZoneIngressOverviews();o.value=t>0}catch(t){o.value=!1,console.error(t)}finally{o.value||(v(),_.value=window.setTimeout(h,b))}}function g(){d.value!==null&&window.clearTimeout(d.value)}function v(){_.value!==null&&window.clearTimeout(_.value)}return(t,J)=>(a(),L(O,null,{default:i(()=>[c(A,{title:w(p)("onboarding.routes.multizone.title")},null,8,["title"]),e(),c(T,null,{default:i(()=>[c(x,null,{header:i(()=>[c(Z,null,{title:i(()=>[e(`
              Add zones
            `)]),_:1})]),content:i(()=>[E,e(),n("p",G,[n("b",null,[e("See "),n("a",{href:w(p)("onboarding.href.docs.install"),target:"_blank"},"the documentation for options to install",8,K),e(".")])]),e(),n("div",null,[n("p",P,[e(`
              Zone status:

              `),s.value?(a(),l("span",U,"Connected")):(a(),l("span",$,"Disconnected"))]),e(),n("p",q,[e(`
              Zone ingress status:

              `),o.value?(a(),l("span",H,"Connected")):(a(),l("span",R,"Disconnected"))]),e(),!o.value||!s.value?(a(),l("div",j,[c(y)])):C("",!0)])]),navigation:i(()=>[c(z,{"next-step":"onboarding-create-mesh","previous-step":"onboarding-configuration-types","should-allow-next":s.value&&o.value},null,8,["should-allow-next"])]),_:1})]),_:1})]),_:1}))}});const ne=V(F,[["__scopeId","data-v-5d5ad5e8"]]);export{ne as default};
