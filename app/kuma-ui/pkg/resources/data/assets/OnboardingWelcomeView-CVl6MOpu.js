import{d as L,I as h,G as x,o as u,p as y,w as l,c as w,H as C,F as I,e as b,a as m,b as a,l as r,t as p,J as V,K as $,m as k,L as q,q as O,_ as B}from"./index-D0CRkaTO.js";import{O as R,a as A,b as E}from"./OnboardingPage-CpLz0Bpy.js";const M=["aria-hidden"],S='<path d="M9.7 18.025L4 12.325L5.425 10.9L9.7 15.175L18.875 6L20.3 7.425L9.7 18.025Z" fill="currentColor"/>',T=L({__name:"CheckIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:h,validator:o=>{if(typeof o=="number"&&o>0)return!0;if(typeof o=="string"){const n=String(o).replace(/px/gi,""),e=Number(n);if(e&&!isNaN(e)&&Number.isInteger(e)&&e>0)return!0}return!1}},as:{type:String,required:!1,default:"span"},staticIds:{type:Boolean,default:!1}},setup(o){const n=o,e=x(()=>{if(typeof n.size=="number"&&n.size>0)return`${n.size}px`;if(typeof n.size=="string"){const i=String(n.size).replace(/px/gi,""),t=Number(i);if(t&&!isNaN(t)&&Number.isInteger(t)&&t>0)return`${t}px`}return h}),c=x(()=>({boxSizing:"border-box",color:n.color,display:n.display,flexShrink:"0",height:e.value,lineHeight:"0",width:e.value,pointerEvents:n.decorative?"none":void 0})),g=i=>{const t={},z=Math.random().toString(36).substring(2,12);return i.replace(/id="([^"]+)"/g,(_,d)=>{const N=`${z}-${d}`;return t[d]=N,`id="${N}"`}).replace(/#([^\s^")]+)/g,(_,d)=>t[d]?`#${t[d]}`:_)},f={"<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;","&":"&amp;"},v=i=>i.replace(/[<>"'&]/g,t=>f[t]),s=`${n.title?`<title data-testid="kui-icon-svg-title">${v(n.title)}</title>`:""}${n.staticIds?S:g(S)}`;return(i,t)=>(u(),y(I(o.as),{"aria-hidden":o.decorative?"true":void 0,class:"kui-icon check-icon","data-testid":"kui-icon-wrapper-check-icon",style:C(c.value),tabindex:o.decorative?"-1":void 0},{default:l(()=>[(u(),w("svg",{"aria-hidden":o.decorative?"true":void 0,"data-testid":"kui-icon-svg-check-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg",innerHTML:s},null,8,M))]),_:1},8,["aria-hidden","style","tabindex"]))}}),H={"data-testid":"kuma-environment"},W={class:"item-status-list-wrapper"},K={class:"item-status-list"},D={class:"circle mr-2"},F=L({__name:"OnboardingWelcomeView",setup(o){return(n,e)=>{const c=b("RouteTitle"),g=b("AppView"),f=b("RouteView");return u(),y(f,{name:"onboarding-welcome-view"},{default:l(({env:v,t:s,can:i})=>[m(c,{title:s("onboarding.routes.welcome.title",{name:s("common.product.name")}),render:!1},null,8,["title"]),e[10]||(e[10]=a()),m(g,null,{default:l(()=>[r("div",null,[m(R,null,{header:l(()=>[m(A,null,{title:l(()=>[a(`
                Welcome to `+p(s("common.product.name")),1)]),description:l(()=>[r("p",null,[a(`
                  Congratulations on downloading `+p(s("common.product.name"))+"! You are just a ",1),e[0]||(e[0]=r("strong",null,"few minutes",-1)),e[1]||(e[1]=a(` away from getting your service mesh fully online.
                `))]),e[4]||(e[4]=a()),r("p",null,[e[2]||(e[2]=a(`
                  We have automatically detected that you are running on `)),r("strong",H,p(s(`common.product.environment.${v("KUMA_ENVIRONMENT")}`)),1),e[3]||(e[3]=a(`.
                `))])]),_:2},1024)]),content:l(()=>[e[6]||(e[6]=r("h2",{class:"text-center"},`
              Let’s get started:
            `,-1)),e[7]||(e[7]=a()),r("div",W,[r("ul",K,[(u(!0),w(V,null,$([{name:`Run ${s("common.product.name")} control plane`,status:!0},{name:"Learn about deployments",status:!1},{name:"Learn about configuration storage",status:!1},...i("use zones")?[{name:"Add zones",status:!1}]:[],{name:"Create the mesh",status:!1},{name:"Add services",status:!1},{name:"Go to the dashboard",status:!1}],t=>(u(),w("li",{key:t.name},[r("span",D,[t.status?(u(),y(k(T),{key:0,size:k(q)},null,8,["size"])):O("",!0)]),a(" "+p(t.name),1)]))),128))])])]),navigation:l(()=>[m(E,{"next-step":"onboarding-deployment-types-view"})]),_:2},1024)])]),_:2},1024)]),_:1})}}}),Z=B(F,[["__scopeId","data-v-e0b0cf7b"]]);export{Z as default};
