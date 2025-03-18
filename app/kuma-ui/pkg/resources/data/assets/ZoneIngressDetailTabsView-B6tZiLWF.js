import{d as q,y as J,r as t,o as d,m as X,w as e,b as o,p as k,ar as S,c as _,q as b,e as s,s as w,P as K,t as g,F as B,v as A,L as P,K as G,_ as H}from"./index-DIs9RbIP.js";const Q={key:0},W=["onSubmit"],Y=["disabled"],ee={key:0},oe={key:0},ne=q({__name:"ZoneIngressDetailTabsView",setup(te){const a=J({eds:!1,xds:!0,clusters:!0,stats:!0,proxy:!0});return(se,n)=>{const I=t("RouteTitle"),L=t("XCopyButton"),z=t("XAction"),x=t("XI18n"),y=t("XCheckbox"),R=t("XAlert"),C=t("DataLoader"),N=t("XDownload"),E=t("XLayout"),M=t("XModal"),V=t("XDisclosure"),U=t("XTeleportTemplate"),$=t("XTabs"),h=t("RouterView"),F=t("AppView"),O=t("DataSource"),Z=t("RouteView");return d(),X(Z,{name:"zone-ingress-detail-tabs-view",params:{zone:"",proxy:""}},{default:e(({route:u,t:r,uri:D})=>[o(O,{src:D(k(S),"/zone-ingress-overviews/:name",{name:u.params.proxy})},{default:e(({data:i,error:j})=>[o(F,{docs:r("zone-ingresses.href.docs"),breadcrumbs:[{to:{name:"zone-cp-list-view"},text:r("zone-cps.routes.item.breadcrumbs")},{to:{name:"zone-cp-detail-view",params:{zone:u.params.zone}},text:u.params.zone},{to:{name:"zone-ingress-list-view",params:{zone:u.params.zone}},text:r("zone-ingresses.routes.item.breadcrumbs")}]},{title:e(()=>[i?(d(),_("h1",Q,[o(L,{text:i.name},{default:e(()=>[o(I,{title:r("zone-ingresses.routes.item.title",{name:i.name})},null,8,["title"])]),_:2},1032,["text"])])):b("",!0)]),actions:e(()=>[o(V,null,{default:e(({expanded:f,toggle:l})=>[o(z,{appearance:"primary",onClick:l},{default:e(()=>n[1]||(n[1]=[s(`
              Download Bundle
            `)])),_:2},1032,["onClick"]),n[6]||(n[6]=s()),f?(d(),X(U,{key:0,to:{name:"modal-layer"}},{default:e(()=>[o(V,null,{default:e(({expanded:v,toggle:T})=>[w("form",{onSubmit:K(T,["prevent"])},[o(M,{title:r("zone-ingresses.routes.item.download.title"),onCancel:l},{"footer-actions":e(()=>[o(E,{type:"separated"},{default:e(()=>[o(N,{onStart:l},{default:e(({download:p})=>[o(C,{variant:"spinner",src:v?D(k(S),"/zone-ingresses/:name/as/tarball/:spec",{name:u.params.proxy,spec:JSON.stringify(a.value)},{cacheControl:"no-cache"}):"",onChange:p,onError:T},{error:e(()=>[o(R,{appearance:"warning","show-icon":""},{default:e(()=>[o(x,{t:"zone-ingresses.routes.item.download.error"})]),_:1})]),_:2},1032,["src","onChange","onError"])]),_:2},1032,["onStart"]),n[4]||(n[4]=s()),o(z,{appearance:"primary",type:"submit",disabled:v||Object.values(a.value).every(p=>!p)},{default:e(()=>[s(g(r("zone-ingresses.routes.item.download.action")),1)]),_:2},1032,["disabled"])]),_:2},1024)]),default:e(()=>[w("fieldset",{disabled:v},[o(x,{path:"zone-ingresses.routes.item.download.description"}),n[3]||(n[3]=s()),w("ul",null,[(d(!0),_(B,null,A(a.value,(p,m)=>(d(),_(B,{key:typeof p},[m!=="eds"?(d(),_("li",ee,[o(y,{modelValue:a.value[m],"onUpdate:modelValue":c=>a.value[m]=c,onChange:c=>{m==="xds"&&!c&&(a.value.eds=!1)}},{default:e(()=>[s(g(r(`zone-ingresses.routes.item.download.options.${m}`)),1)]),_:2},1032,["modelValue","onUpdate:modelValue","onChange"]),n[2]||(n[2]=s()),m==="xds"?(d(),_("ul",oe,[w("li",null,[o(y,{modelValue:a.value.eds,"onUpdate:modelValue":n[0]||(n[0]=c=>a.value.eds=c),disabled:!a.value.xds},{default:e(()=>[s(g(r("zone-ingresses.routes.item.download.options.eds")),1)]),_:2},1032,["modelValue","disabled"])])])):b("",!0)])):b("",!0)],64))),128))])],8,Y),n[5]||(n[5]=s())]),_:2},1032,["title","onCancel"])],40,W)]),_:2},1024)]),_:2},1024)):b("",!0)]),_:2},1024)]),default:e(()=>[n[8]||(n[8]=s()),n[9]||(n[9]=s()),o(C,{data:[i],errors:[j]},{default:e(()=>{var f;return[o($,{selected:(f=u.child())==null?void 0:f.name,"data-testid":"zone-ingress-tabs"},P({_:2},[A(u.children,({name:l})=>({name:`${l}-tab`,fn:e(()=>[o(z,{to:{name:l}},{default:e(()=>[s(g(r(`zone-ingresses.routes.item.navigation.${l}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),n[7]||(n[7]=s()),o(h,null,{default:e(l=>[(d(),X(G(l.Component),{networking:i==null?void 0:i.zoneIngress.networking,data:i},null,8,["networking","data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}}),re=H(ne,[["__scopeId","data-v-beba8061"]]);export{re as default};
