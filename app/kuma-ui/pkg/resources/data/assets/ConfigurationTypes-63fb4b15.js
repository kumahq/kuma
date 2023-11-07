import{O as h,a as G,b as x}from"./OnboardingPage-2d546b3e.js";import{d as C,k as R,Q as T,R as w,S as P,m as K,P as k,a as r,o as u,b as d,w as e,e as o,f as t,p,G as M,_ as O}from"./index-70fb4e48.js";const N={class:"graph-list mb-6"},U={class:"radio-button-group"},A=C({__name:"ConfigurationTypes",setup(B){const i=R(),m=T(),c=w(),_={postgres:P(),memory:c,kubernetes:m},n=K(i("KUMA_STORE_TYPE")),g=k(()=>_[n.value]);return(z,a)=>{const v=r("RouteTitle"),l=r("KRadio"),f=r("AppView"),b=r("RouteView");return u(),d(b,{name:"onboarding-configuration-types"},{default:e(({can:y,t:V})=>[o(v,{title:V("onboarding.routes.configuration-types.title")},null,8,["title"]),t(),o(f,null,{default:e(()=>[o(h,{"with-image":""},{header:e(()=>[o(G,null,{title:e(()=>[t(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[p("div",N,[(u(),d(M(g.value)))]),t(),p("div",U,[o(l,{modelValue:n.value,"onUpdate:modelValue":a[0]||(a[0]=s=>n.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[t(`
              Kubernetes
            `)]),_:1},8,["modelValue"]),t(),o(l,{modelValue:n.value,"onUpdate:modelValue":a[1]||(a[1]=s=>n.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[t(`
              Postgres
            `)]),_:1},8,["modelValue"]),t(),o(l,{modelValue:n.value,"onUpdate:modelValue":a[2]||(a[2]=s=>n.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[t(`
              Memory
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[o(x,{"next-step":y("use zones")?"onboarding-multi-zone":"onboarding-create-mesh","previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const H=O(A,[["__scopeId","data-v-d26eecda"]]);export{H as default};
