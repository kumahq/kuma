import{O as w,a as h,b as C}from"./OnboardingPage-dc8bc63e.js";import{d as x,u as G,I as O,J as T,K,B as R,H as P,a as r,o as i,b as u,w as e,e as o,f as n,m as d,C as M,_ as k}from"./index-f5266944.js";const B={class:"graph-list mb-6"},N={class:"radio-button-group"},U=x({__name:"OnboardingConfigurationTypesView",setup(A){const p=G(),m=O(),_=T(),c={postgres:K(),memory:_,kubernetes:m},t=R(p("KUMA_STORE_TYPE")),g=P(()=>c[t.value]);return(z,a)=>{const v=r("RouteTitle"),l=r("KRadio"),b=r("AppView"),f=r("RouteView");return i(),u(f,{name:"onboarding-configuration-types-view"},{default:e(({can:V,t:y})=>[o(v,{title:y("onboarding.routes.configuration-types.title"),render:!1},null,8,["title"]),n(),o(b,null,{default:e(()=>[o(w,{"with-image":""},{header:e(()=>[o(h,null,{title:e(()=>[n(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[d("div",B,[(i(),u(M(g.value)))]),n(),d("div",N,[o(l,{modelValue:t.value,"onUpdate:modelValue":a[0]||(a[0]=s=>t.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[n(`
              Kubernetes
            `)]),_:1},8,["modelValue"]),n(),o(l,{modelValue:t.value,"onUpdate:modelValue":a[1]||(a[1]=s=>t.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[n(`
              Postgres
            `)]),_:1},8,["modelValue"]),n(),o(l,{modelValue:t.value,"onUpdate:modelValue":a[2]||(a[2]=s=>t.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[n(`
              Memory
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[o(C,{"next-step":V("use zones")?"onboarding-multi-zone-view":"onboarding-create-mesh-view","previous-step":"onboarding-deployment-types-view"},null,8,["next-step"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const D=k(U,[["__scopeId","data-v-12112a8a"]]);export{D as default};
