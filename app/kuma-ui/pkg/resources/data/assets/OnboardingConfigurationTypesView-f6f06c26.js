import{O as w,a as h,b as G}from"./OnboardingPage-1bfbbaa9.js";import{d as O,k as x,O as C,Q as R,R as T,m as K,N as P,a as r,o as i,b as u,w as e,e as o,f as n,p as d,G as k,_ as M}from"./index-81fc4a03.js";const N={class:"graph-list mb-6"},U={class:"radio-button-group"},A=O({__name:"OnboardingConfigurationTypesView",setup(B){const p=x(),m=C(),_=R(),c={postgres:T(),memory:_,kubernetes:m},t=K(p("KUMA_STORE_TYPE")),g=P(()=>c[t.value]);return(z,a)=>{const v=r("RouteTitle"),l=r("KRadio"),b=r("AppView"),f=r("RouteView");return i(),u(f,{name:"onboarding-configuration-types-view"},{default:e(({can:V,t:y})=>[o(v,{title:y("onboarding.routes.configuration-types.title"),render:!1},null,8,["title"]),n(),o(b,null,{default:e(()=>[o(w,{"with-image":""},{header:e(()=>[o(h,null,{title:e(()=>[n(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[d("div",N,[(i(),u(k(g.value)))]),n(),d("div",U,[o(l,{modelValue:t.value,"onUpdate:modelValue":a[0]||(a[0]=s=>t.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[n(`
              Kubernetes
            `)]),_:1},8,["modelValue"]),n(),o(l,{modelValue:t.value,"onUpdate:modelValue":a[1]||(a[1]=s=>t.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[n(`
              Postgres
            `)]),_:1},8,["modelValue"]),n(),o(l,{modelValue:t.value,"onUpdate:modelValue":a[2]||(a[2]=s=>t.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[n(`
              Memory
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[o(G,{"next-step":V("use zones")?"onboarding-multi-zone-view":"onboarding-create-mesh-view","previous-step":"onboarding-deployment-types-view"},null,8,["next-step"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const I=M(A,[["__scopeId","data-v-12112a8a"]]);export{I as default};
