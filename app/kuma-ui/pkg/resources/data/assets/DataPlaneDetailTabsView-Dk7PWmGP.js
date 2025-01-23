import{d as W,x as z,r as o,o as d,q as v,w as t,b as a,p as P,V as S,Q as A,e as n,m as f,W as H,c as w,K as C,L as x,t as h,s as y,G as Y,_ as Z}from"./index-CKQWVGYP.js";const ee=["onSubmit"],te=["disabled"],ae={key:0},oe={key:0},ne=W({__name:"DataPlaneDetailTabsView",props:{mesh:{}},setup(N){const U=N,s=z({eds:!1,xds:!0,clusters:!0,stats:!0,dataplane:!0,policies:!0}),E=k=>async e=>{const p=document.createElement("a");p.download=e.name,p.href=e.url,setTimeout(()=>{window.URL.revokeObjectURL(p.href)},6e4),await Promise.resolve(),p.click(),await Promise.resolve(),k()};return(k,e)=>{const p=o("RouteTitle"),I=o("XCopyButton"),V=o("XAction"),D=o("XI18n"),T=o("XCheckbox"),M=o("XAlert"),g=o("DataLoader"),O=o("XLayout"),$=o("XModal"),B=o("XDisclosure"),j=o("XTeleportTemplate"),q=o("XTabs"),F=o("RouterView"),G=o("AppView"),J=o("DataSource"),K=o("RouteView");return d(),v(K,{name:"data-plane-detail-tabs-view",params:{mesh:"",dataPlane:""}},{default:t(({route:l,t:i,uri:L})=>[a(J,{src:L(P(S),"/meshes/:mesh/dataplane-overviews/:name",{mesh:l.params.mesh,name:l.params.dataPlane})},{default:t(({data:m,error:Q})=>[a(G,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:l.params.mesh}},text:l.params.mesh},{to:{name:"data-plane-list-view",params:{mesh:l.params.mesh}},text:i("data-planes.routes.item.breadcrumbs")}]},A({actions:t(()=>[a(B,null,{default:t(({expanded:b,toggle:r})=>[a(V,{appearance:"primary",onClick:r},{default:t(()=>e[1]||(e[1]=[n(`
              Download Bundle
            `)])),_:2},1032,["onClick"]),e[6]||(e[6]=n()),b?(d(),v(j,{key:0,to:{name:"modal-layer"}},{default:t(()=>[a(B,null,{default:t(({expanded:X,toggle:R})=>[f("form",{onSubmit:H(R,["prevent"])},[a($,{title:i("data-planes.routes.item.download.title"),onCancel:r},{"footer-actions":t(()=>[a(O,{type:"separated"},{default:t(()=>[(d(!0),w(C,null,x([E(r)],u=>(d(),v(g,{key:typeof u,variant:"spinner",src:X?L(P(S),"/meshes/:mesh/dataplanes/:name/as/tarball/:spec",{mesh:l.params.mesh,name:l.params.dataPlane,spec:JSON.stringify(s.value)},{cacheControl:"no-cache"}):"",onChange:u,onError:R},{error:t(()=>[a(M,{appearance:"warning","show-icon":""},{default:t(()=>[a(D,{t:"data-planes.routes.item.download.error"})]),_:1})]),_:2},1032,["src","onChange","onError"]))),128)),e[4]||(e[4]=n()),a(V,{appearance:"primary",type:"submit",disabled:X||Object.values(s.value).every(u=>!u)},{default:t(()=>[n(h(i("data-planes.routes.item.download.action")),1)]),_:2},1032,["disabled"])]),_:2},1024)]),default:t(()=>[f("fieldset",{disabled:X},[a(D,{path:"data-planes.routes.item.download.description"}),e[3]||(e[3]=n()),f("ul",null,[(d(!0),w(C,null,x(s.value,(u,c)=>(d(),w(C,{key:typeof u},[c!=="eds"?(d(),w("li",ae,[a(T,{modelValue:s.value[c],"onUpdate:modelValue":_=>s.value[c]=_,onChange:_=>{c==="xds"&&!_&&(s.value.eds=!1)}},{default:t(()=>[n(h(i(`data-planes.routes.item.download.options.${c}`)),1)]),_:2},1032,["modelValue","onUpdate:modelValue","onChange"]),e[2]||(e[2]=n()),c==="xds"?(d(),w("ul",oe,[f("li",null,[a(T,{modelValue:s.value.eds,"onUpdate:modelValue":e[0]||(e[0]=_=>s.value.eds=_),disabled:!s.value.xds},{default:t(()=>[n(h(i("data-planes.routes.item.download.options.eds")),1)]),_:2},1032,["modelValue","disabled"])])])):y("",!0)])):y("",!0)],64))),128))])],8,te),e[5]||(e[5]=n())]),_:2},1032,["title","onCancel"])],40,ee)]),_:2},1024)]),_:2},1024)):y("",!0)]),_:2},1024)]),default:t(()=>[e[8]||(e[8]=n()),e[9]||(e[9]=n()),a(g,{data:[m],errors:[Q]},{default:t(()=>{var b;return[a(q,{selected:(b=l.child())==null?void 0:b.name},A({_:2},[x(l.children,({name:r})=>({name:`${r}-tab`,fn:t(()=>[a(V,{to:{name:r}},{default:t(()=>[n(h(i(`data-planes.routes.item.navigation.${r}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),e[7]||(e[7]=n()),a(F,null,{default:t(({Component:r})=>[(d(),v(Y(r),{data:m,networking:m==null?void 0:m.dataplane.networking,mesh:U.mesh},null,8,["data","networking","mesh"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},[m?{name:"title",fn:t(()=>[f("h1",null,[a(I,{text:m.name},{default:t(()=>[a(p,{title:i("data-planes.routes.item.title",{name:m.name})},null,8,["title"])]),_:2},1032,["text"])])]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}}),le=Z(ne,[["__scopeId","data-v-97116bf6"]]);export{le as default};
